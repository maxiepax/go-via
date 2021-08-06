package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/maxiepax/go-via/secrets"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

var defaultks = `
# Accept the VMware End User License Agreement
vmaccepteula

# Set the root password for the DCUI and Tech Support Mode
rootpw {{ .password }}

{{ if .erasedisks }}
# Remove ALL partitions
clearpart --overwritevmfs --alldrives {{ end }}

{{ if .bootdisk }}
install --disk=/vmfs/devices/disks/{{.bootdisk}} --overwritevmfs --novmfsondisk
{{ else }}
# Install on the first local disk available on machine
install --overwritevmfs --novmfsondisk --firstdisk="localesx,usb,ahci,vmw_ahci,VMware"
{{ end }}

# Set the network to static on the first network adapter
network --bootproto=static --ip={{ .ip }} --gateway={{ .gateway }} --netmask={{ .netmask }} --nameserver={{ .dns }} --hostname={{ .hostname }} --device=vmnic0

reboot
`

//func Ks(c *gin.Context) {
func Ks(key string) func(c *gin.Context) {
	return func(c *gin.Context) {
		fmt.Println(key)
		var item models.Address
		host, _, _ := net.SplitHostPort(c.Request.RemoteAddr)

		if res := db.DB.Preload(clause.Associations).Where("ip = ?", host).First(&item); res.Error != nil {
			Error(c, http.StatusInternalServerError, res.Error) // 500
			return
		}

		options := models.GroupOptions{}
		json.Unmarshal(item.Group.Options, &options)

		if reimage := db.DB.Model(&item).Where("ip = ?", host).Update("reimage", false); reimage.Error != nil {
			Error(c, http.StatusInternalServerError, reimage.Error) // 500
			return
		}

		laddrport, ok := c.Request.Context().Value(http.LocalAddrContextKey).(net.Addr)
		if !ok {
			logrus.WithFields(logrus.Fields{
				"interface": "could not determine the local interface used to apply to ks.cfgs postconfig callback",
			}).Debug("ks")
		}

		logrus.Info("Disabling re-imaging for host to avoid re-install looping")

		//convert netmask from bit to long format.
		nm := net.CIDRMask(item.Pool.Netmask, 32)
		netmask := ipv4MaskString(nm)

		//decrypt the password
		decryptedPassword := secrets.Decrypt(item.Group.Password, key)

		//cleanup data to allow easier custom templating
		data := map[string]interface{}{
			"password":   decryptedPassword,
			"ip":         item.IP,
			"gateway":    item.Pool.Gateway,
			"dns":        item.Group.DNS,
			"hostname":   item.Hostname,
			"netmask":    netmask,
			"via_server": laddrport,
			"erasedisks": options.EraseDisks,
			"bootdisk":   options.BootDisk,
		}

		// check if default ks has been overridden.
		ks := defaultks
		if item.Group.Ks != "" {
			ks = item.Group.Ks
		}

		t, err := template.New("").Parse(ks)
		if err != nil {
			logrus.Info(err)
			return
		}
		err = t.Execute(c.Writer, data)
		if err != nil {
			logrus.Info(err)
			return
		}

		logrus.Info("Served ks.cfg file")
		logrus.WithFields(logrus.Fields{
			"id":      item.ID,
			"ip":      item.IP,
			"host":    item.Hostname,
			"message": "served ks.cfg file",
		}).Info("ks")
		logrus.WithFields(logrus.Fields{
			"id":           item.ID,
			"percentage":   50,
			"progresstext": "kickstart",
		}).Info("progress")
		item.Progress = 50
		item.Progresstext = "kickstart"
		db.DB.Save(&item)

		go ProvisioningWorker(item, key)

		logrus.Info("Started worker")
	}
}

func ipv4MaskString(m []byte) string {
	if len(m) != 4 {
		panic("ipv4Mask: len must be 4 bytes")
	}

	return fmt.Sprintf("%d.%d.%d.%d", m[0], m[1], m[2], m[3])
}
