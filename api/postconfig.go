package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	ca "github.com/maxiepax/go-via/crypto"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/maxiepax/go-via/secrets"
	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/govc/host/esxcli"
	"github.com/vmware/govmomi/vim25/types"
	"gorm.io/gorm/clause"
)

func PostConfig(key string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var item models.Address
		host, _, _ := net.SplitHostPort(c.Request.RemoteAddr)

		if res := db.DB.Preload(clause.Associations).Where("ip = ?", host).First(&item); res.Error != nil {
			Error(c, http.StatusInternalServerError, res.Error) // 500
			return
		}

		c.JSON(http.StatusOK, item) // 200

		logrus.Info("ks config done!")

		go ProvisioningWorker(item, key)
	}
}

func PostConfigID(key string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var item models.Address

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Error(c, http.StatusBadRequest, err) // 400
			return
		}

		if res := db.DB.Preload(clause.Associations).Where("id = ?", id).First(&item); res.Error != nil {
			Error(c, http.StatusInternalServerError, res.Error) // 500
			return
		}

		c.JSON(http.StatusOK, item) // 200

		logrus.Info("Manual PostConfig of host" + item.Hostname + "started!")

		go ProvisioningWorker(item, key)
	}
}

func ProvisioningWorker(item models.Address, key string) {

	//create empty model and load it with the json content from database
	options := models.GroupOptions{}
	json.Unmarshal(item.Group.Options, &options)
	logrus.WithFields(logrus.Fields{
		"Started worker for ": item.Hostname,
	}).Debug("host")

	// decrypt login password
	decryptedPassword := secrets.Decrypt(item.Group.Password, key)

	// connection info
	url := &url.URL{
		Scheme: "https",
		Host:   item.IP,
		Path:   "sdk",
		User:   url.UserPassword("root", decryptedPassword),
	}

	logrus.WithFields(logrus.Fields{
		"id":           item.ID,
		"percentage":   75,
		"progresstext": "customization",
	}).Info("progress")
	item.Progress = 75
	item.Progresstext = "customization"
	db.DB.Save(&item)

	// ensure that host has enough time to boot, and for SOAP API to respond.
	var c *govmomi.Client
	var err error
	ctx := context.Background()
	i := 1
	timeout := 360
	for {
		if i > timeout {
			logrus.WithFields(logrus.Fields{
				"IP":     item.IP,
				"status": "timeout exceeded, failing postconfig",
			}).Info("postconfig")
			return
		}

		if res := db.DB.First(&item, item.ID); res.Error != nil {
			logrus.WithFields(logrus.Fields{
				"IP":  item.IP,
				"err": res.Error,
			}).Error("postconfig failed to read state")
			return
		}

		if !item.Reimage {
			logrus.WithFields(logrus.Fields{
				"IP": item.IP,
			}).Error("postconfig terminated")
			return
		}

		c, err = govmomi.NewClient(ctx, url, true)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"IP":        item.IP,
				"status":    "Hosts SOAP API not ready yet, retrying",
				"retry":     i,
				"retry max": timeout,
			}).Info("postconfig")
			logrus.WithFields(logrus.Fields{
				"IP":        item.IP,
				"status":    "Hosts SOAP API not ready yet, retrying",
				"retry":     i,
				"retry max": timeout,
				"err":       err,
			}).Debug("postconfig")
			i += 1
			<-time.After(time.Second * 10)
			continue
		}
		break
	}

	// since we're always going to be talking directly to the host, dont asume connection through vCenter.
	host, err := find.NewFinder(c.Client).DefaultHostSystem(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"postconfig": err,
		}).Info(item.IP)
	}
	e, err := esxcli.NewExecutor(c.Client, host)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"postconfig": err,
		}).Info(item.IP)
	}

	if item.Domain != "" {
		search := strings.Fields("network ip dns search add -d")
		search = append(search, item.Domain)
		_, err := e.Run(search)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":            item.IP,
			"search domain": item.Domain,
		}).Info("postconfig")

		hd := string(item.Hostname + "." + item.Domain)
		fqdn := strings.Fields("system hostname set --fqdn")
		fqdn = append(fqdn, hd)
		_, err = e.Run(fqdn)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":   item.IP,
			"fqdn": hd,
		}).Info("postconfig")
	}

	if item.Group.NTP != "" {
		//configure ntpd
		cmd := strings.Fields("system ntp set")
		for _, k := range strings.Split(item.Group.NTP, ",") {
			cmd = append(cmd, "--server", string(k))
		}
		_, err := e.Run(cmd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"ntpd": item.Group.NTP,
		}).Info(item.IP)
		//enable ntpd
		_, err = e.Run(strings.Fields("system ntp set --enabled true"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":   item.IP,
			"ntpd": "Service enabled",
		}).Info("postconfig")

		s, err := host.ConfigManager().ServiceSystem(ctx)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}

		err = s.UpdatePolicy(ctx, "ntpd", string(types.HostServicePolicyOn))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":   item.IP,
			"ntpd": "Startup Policy -> Start and stop with host",
		}).Info("postconfig")

		err = s.Start(ctx, "ntpd")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":   item.IP,
			"ntpd": "Service started",
		}).Info("postconfig")
	}

	if item.Group.Syslog != "" {
		//configure Syslog and modify firewall to allow syslog.
		cmd := strings.Fields("system syslog config set --loghost=" + item.Group.Syslog)
		_, err := e.Run(cmd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		_, err = e.Run(strings.Fields("system syslog reload"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		_, err = e.Run(strings.Fields("network firewall ruleset set --ruleset-id=syslog --enabled=true"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		_, err = e.Run(strings.Fields("network firewall refresh"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":     item.IP,
			"syslog": "syslog configured",
		}).Info("postconfig")
	}

	if options.SSH {
		s, err := host.ConfigManager().ServiceSystem(ctx)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}

		err = s.UpdatePolicy(ctx, "TSM-SSH", string(types.HostServicePolicyOn))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":  item.IP,
			"ssh": "Startup Policy -> Start and stop with host",
		}).Info("postconfig")

		err = s.Start(ctx, "TSM-SSH")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":  item.IP,
			"ssh": "Service started",
		}).Info("postconfig")
	}

	if options.SSH {
		//Suppress any warnings that ESXi Console or SSH are enabled
		cmd := strings.Fields("system settings advanced set -o /UserVars/SuppressShellWarning -i 1")
		_, err := e.Run(cmd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":  item.IP,
			"ssh": "suppressing shell warnings",
		}).Info("postconfig")
	}

	if item.Group.Vlan != "" {
		//if vlan is set, configure the "VM Network" portgroup with the same vlanid.

		cmd := strings.Fields("network vswitch standard portgroup set --vlan-id " + item.Group.Vlan)
		cmd = append(cmd, "-p")
		cmd = append(cmd, "VM Network")

		_, err := e.Run(cmd)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":   item.IP,
			"vlan": "VM Network vlan-id : " + item.Group.Vlan,
		}).Info("postconfig")
	}

	if options.Certificate {
		//create directory
		os.MkdirAll("./cert/"+item.Hostname+"."+item.Domain, os.ModePerm)
		//create certificate
		ca.CreateCert("./cert/"+item.Hostname+"."+item.Domain, "rui", item.Hostname+"."+item.Domain)

		//post to esxi host https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.security.doc/GUID-43B7B817-C58F-4C6F-AF3D-9F1D52B116A0.html
		crt, err := os.Open("./cert/" + item.Hostname + "." + item.Domain + "/rui.crt")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": "couldn't find the .crt file",
			}).Info(item.IP)
		}
		defer crt.Close()

		key, err := os.Open("./cert/" + item.Hostname + "." + item.Domain + "/rui.key")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": "couldn't find the .key file",
			}).Info(item.IP)
		}
		defer key.Close()

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		putRequest("https://"+item.IP+"/host/ssl_cert", crt, "root", decryptedPassword)
		putRequest("https://"+item.IP+"/host/ssl_key", key, "root", decryptedPassword)

		// set the host into maintenanace mode
		cmd := strings.Fields("system maintenanceMode set -e true")
		_, err = e.Run(cmd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":          item.IP,
			"certificate": "set host into maintenance mode",
		}).Info("postconfig")

		// reboot the host
		cmd = strings.Fields("system shutdown reboot -r certificate")
		_, err = e.Run(cmd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":          item.IP,
			"certificate": "rebooting host to activate new certificates",
		}).Info("postconfig")

		logrus.WithFields(logrus.Fields{
			"id":           item.ID,
			"percentage":   90,
			"progresstext": "rebooting host",
		}).Info("progress")
		item.Progress = 90
		item.Progresstext = "rebooting host"
		db.DB.Save(&item)

		// wait for the SOAP API to come back
		time.Sleep(15 * time.Second)
		for {
			if i > timeout {
				logrus.WithFields(logrus.Fields{
					"IP":     item.IP,
					"status": "timeout exceeded, failing postconfig",
				}).Info("postconfig")
				return
			}
			c, err = govmomi.NewClient(ctx, url, true)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"IP":        item.IP,
					"status":    "Hosts SOAP API not ready yet, retrying",
					"retry":     i,
					"retry max": timeout,
				}).Info("postconfig")
				logrus.WithFields(logrus.Fields{
					"IP":        item.IP,
					"status":    "Hosts SOAP API not ready yet, retrying",
					"retry":     i,
					"retry max": timeout,
					"err":       err,
				}).Debug("postconfig")
				i += 1
				<-time.After(time.Second * 10)
				continue
			}
			break
		}

		// re-authenticate and create new session since last reboot.
		host, err := find.NewFinder(c.Client).DefaultHostSystem(ctx)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		e, err := esxcli.NewExecutor(c.Client, host)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}

		// bring host out of maintenanace mode
		cmd = strings.Fields("system maintenanceMode set -e false")
		_, err = e.Run(cmd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"IP":          item.IP,
			"certificate": "remove host from maintenance mode",
		}).Info("postconfig")

	}

	logrus.WithFields(logrus.Fields{
		"postconfig": "postconfig completed",
	}).Info(item.IP)

	logrus.WithFields(logrus.Fields{
		"id":           item.ID,
		"percentage":   100,
		"progresstext": "completed",
	}).Info("progress")
	item.Progress = 100
	item.Progresstext = "completed"
	db.DB.Save(&item)

}

func putRequest(url string, data io.Reader, username string, password string) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, data)
	req.SetBasicAuth(username, password)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"postconfig": err,
		}).Info("")
	}
	_, err = client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"postconfig": err,
		}).Info("")
	}
}
