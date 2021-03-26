package api

import (
	"net"
	"net/http"
	"text/template"

	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

var defaultks = `
#
# Sample scripted installation file
#

# Accept the VMware End User License Agreement
vmaccepteula

{{ if ne .Group.Password "" }}
# Set the root password for the DCUI and Tech Support Mode
rootpw {{ .Group.Password }}
{{ end }}


# Install on the first local disk available on machine
install --firstdisk --overwritevmfs

# Set the network to static on the first network adapter
network --bootproto=static --ip={{ .IP }} --gateway={{ .Pool.Gateway }} --netmask=255.255.255.0 --nameserver={{ .Group.DNS }} --hostname={{ .Hostname }} --device=vmnic0

%firstboot --interpreter=busybox

sleep 20
esxcli network ip dns search add --domain={{ .Domain }}
esxcli network ip dns server add --server=192.168.1.1

# enable & start remote ESXi Shell  (SSH)
vim-cmd hostsvc/enable_ssh
vim-cmd hostsvc/start_ssh
#Suppress shell warning
esxcli system settings advanced set -o /UserVars/SuppressShellWarning -i 1

# NTP Configuration (thanks to http://www.virtuallyghetto.com)
cat > /etc/ntp.conf << __NTP_CONFIG__
restrict default kod nomodify notrap noquerynopeer
restrict 127.0.0.1
server 0.fr.pool.ntp.org
server 1.fr.pool.ntp.org
__NTP_CONFIG__
 
/sbin/chkconfig ntpd on

# A sample post-install script
%post --interpreter=python --ignorefailure=true
import time
stampFile = open('/finished.stamp', mode='w')
stampFile.write( time.asctime() )
`

func Ks(c *gin.Context) {
	var item models.Address
	host, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	//if res := db.DB.Where("ip = ?", host).First(&item); res.Error != nil {
	if res := db.DB.Preload(clause.Associations).Where("ip = ?", host).First(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	if reserved := db.DB.Model(&item).Where("ip = ?", host).Update("reserved", false); reserved.Error != nil {
		Error(c, http.StatusInternalServerError, reserved.Error) // 500
		return
	}

	logrus.Info("Disabling re-imaging for host to avoid re-install looping")

	ntp := strings.Split(ntp, ",")
	spew.Dump(ntp)

	c.JSON(http.StatusOK, item) // 200

	// check if default ks has been overridden.
	ks := defaultks
	if item.Group.Ks != "" {
		ks = item.Group.Ks
	}

	t, err := template.New("").Parse(ks)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = t.Execute(c.Writer, item)
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.Info("Served ks.cfg file")
}
