package api

import (
	"fmt"
	"net"
	"net/http"
	"text/template"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

var defaultks = `
# Accept the VMware End User License Agreement
vmaccepteula

{{ if ne .model.Group.Password "" }}
# Set the root password for the DCUI and Tech Support Mode
rootpw {{ .model.Group.Password }}{{ end }}

# Install on the first local disk available on machine
install --firstdisk --overwritevmfs

# Set the network to static on the first network adapter
network --bootproto=static --ip={{ .model.IP }} --gateway={{ .model.Pool.Gateway }} --netmask={{ .netmask }} --nameserver={{ .model.Group.DNS }} --hostname={{ .model.Hostname }} --device=vmnic0

%firstboot --interpreter=busybox

sleep 20
esxcli network ip dns search add --domain={{ .model.Domain }}
esxcli network ip dns server add --server=192.168.1.1

# enable & start remote ESXi Shell  (SSH)
vim-cmd hostsvc/enable_ssh
vim-cmd hostsvc/start_ssh
#Suppress shell warning
esxcli system settings advanced set -o /UserVars/SuppressShellWarning -i 1

# NTP Configuration (thanks to http://www.virtuallyghetto.com)
cat > /etc/ntp.conf << __NTP_CONFIG__
restrict default nomodify notrap nopeer noquery
restrict 127.0.0.1
restrict -6 ::1
driftfile /etc/ntp.drift
{{ range .ntp }}
server {{ . }}{{ end }}
__NTP_CONFIG__
 
#enable ntpd
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

	ntp := strings.Split(item.Group.NTP, ",")
	netmask := net.CIDRMask(item.pool.netmask, 32)
	netmask = ipv4MaskString(netmask)

	data := map[string]interface{}{
		"model":   item,
		"ntp":     ntp,
		"netmask": netmask,
	}

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
	err = t.Execute(c.Writer, data)
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.Info("Served ks.cfg file")
}

func ipv4MaskString(m []byte) string {
	if len(m) != 4 {
		panic("ipv4Mask: len must be 4 bytes")
	}

	return fmt.Sprintf("%d.%d.%d.%d", m[0], m[1], m[2], m[3])
}
