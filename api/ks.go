package api

import (
	"net/http"
	"text/template"

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

# Set the network to DHCP on the first network adapter
network --bootproto=dhcp --device=vmnic0

# A sample post-install script
%post --interpreter=python --ignorefailure=true
import time
stampFile = open('/finished.stamp', mode='w')
stampFile.write( time.asctime() )
`

func Ks(c *gin.Context) {
	var item models.Address
	//host, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	//if res := db.DB.Where("ip = ?", host).First(&item); res.Error != nil {
	if res := db.DB.Preload(clause.Associations).First(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}
	//c.JSON(http.StatusOK, item) // 200

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
}
