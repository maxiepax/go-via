package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/maxiepax/go-via/db"

	"github.com/maxiepax/go-via/models"
	"github.com/maxiepax/go-via/secrets"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
	//"github.com/davecgh/go-spew/spew"
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
install --disk=/vmfs/devices/disks/{{.bootdisk}} --overwritevmfs --novmfsondisk {{ if not .legacycpu }} --forceunsupportedinstall {{ end }}
{{ else }}
# Install on the first local disk available on machine
install --overwritevmfs {{ if not .createvmfs }} --novmfsondisk {{ end }} --firstdisk="localesx,usb,ahci,vmw_ahci,VMware" --forceunsupportedinstall
{{ end }}

# Set the network to static on the first network adapter
network --bootproto=static --ip={{ .ip }} --gateway={{ .gateway }} --netmask={{ .netmask }} {{if .dns}}--nameserver={{ .dns }} {{end}} --hostname={{ .hostname }} --device={{ .mac }} {{if .vlan}} --vlanid={{.vlan}} {{end}}

reboot

%firstboot --interpreter=busybox

# Configure NTP
{{ if .ntp }}
esxcli system ntp set -e true -s {{ .ntp }}
{{ end }}


# Configure Domain Search
{{ if .domain }}
esxcli network ip dns search add -d {{ .domain }}
{{ end }}

# Configure FQDN
esxcli system hostname set --fqdn {{ .fqdn }}

# Enable SSH
{{ if .ssh }}
vim-cmd hostsvc/enable_ssh
vim-cmd hostsvc/start_ssh
system settings advanced set -o /UserVars/SuppressShellWarning -i 1
{{ end }}

# Syslog
{{ if .syslog }}
esxcli system syslog config set --loghost={{ .syslog }}
esxcli system syslog reload
esxcli network firewall ruleset set --ruleset-id=syslog --enabled=true
esxcli network firewall refresh
{{ end }}

#vSwitch0
{{ if .vlan }}
esxcli network vswitch standard portgroup set --vlan-id {{.vlan}}
{{ end }}

# Ensure TLS certificate matches ESXi FQDN
/sbin/generate-certificates
/etc/init.d/hostd restart && /etc/init.d/vpxa restart && /etc/init.d/rhttpproxy restart
`

// func Ks(c *gin.Context) {
func Ks(key string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var item models.Host
		host, _, _ := net.SplitHostPort(c.Request.RemoteAddr)

		if res := db.DB.Preload(clause.Associations).Where("ip = ?", host).First(&item); res.Error != nil {
			Error(c, http.StatusInternalServerError, res.Error) // 500
			return
		}

		options := models.GroupOptions{}
		err := json.Unmarshal(item.Group.Options, &options)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": "couldn't unmarshal group options",
			}).Debug(item.IP)
			return
		}

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

		//split NTP
		ntp := strings.Fields("esxcli system ntp set")
		for _, k := range strings.Split(item.Group.NTP, ",") {
			ntp = append(ntp, "--server", string(k))
		}

		//cleanup data to allow easier custom templating
		data := map[string]interface{}{
			"password":   decryptedPassword,
			"ip":         item.IP,
			"mac":        item.Mac,
			"gateway":    item.Pool.Gateway,
			"dns":        item.Group.DNS,
			"ntp":        ntp,
			"hostname":   item.Hostname,
			"domain":     item.Domain,
			"fqdn":       item.Hostname + "." + item.Domain,
			"netmask":    netmask,
			"via_server": laddrport,
			"erasedisks": options.EraseDisks,
			"ssh":        options.SSH,
			"syslog":     item.Group.Syslog,
			"bootdisk":   item.Group.BootDisk,
			"vlan":       item.Group.Vlan,
			"createvmfs": options.CreateVMFS,
			"legacycpu":  options.AllowLegacyCPU,
		}

		ks := defaultks

		// check if default ks has been overridden.
		if item.Ks != "" {
			dec, _ := base64.StdEncoding.DecodeString(item.Ks)
			ks = string(dec)
			logrus.WithFields(logrus.Fields{
				"custom host ks": ks,
			}).Debug("ks")
		} else if item.Group.Ks != "" {
			dec, _ := base64.StdEncoding.DecodeString(item.Group.Ks)
			ks = string(dec)
			logrus.WithFields(logrus.Fields{
				"custom group ks": ks,
			}).Debug("ks")
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

		//debug ks.cfg output
		//spew.Dump(t.Execute(os.Stdout, data))

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
