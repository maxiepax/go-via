package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/govc/host/esxcli"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
	"gorm.io/gorm/clause"
)

func PostConfig(c *gin.Context) {
	var item models.Address
	host, _, _ := net.SplitHostPort(c.Request.RemoteAddr)

	if res := db.DB.Preload(clause.Associations).Where("ip = ?", host).First(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusOK, item) // 200

	logrus.Info("ks config done!")

	go ProvisioningWorker(item)

}

func PostConfigID(c *gin.Context) {
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

	c.JSON(http.StatusOK, "OK") // 200

	logrus.Info("Manual PostConfig of host" + item.Hostname + "started!")

	go ProvisioningWorker(item)

}

func ProvisioningWorker(item models.Address) {

	//create empty model and load it with the json content from database
	options := models.GroupOptions{}
	json.Unmarshal(item.Group.Options, &options)
	logrus.WithFields(logrus.Fields{
		"Started worker for ": item.Hostname,
	}).Debug("host")

	// connection info
	url := &url.URL{
		Scheme: "https",
		Host:   item.IP,
		Path:   "sdk",
		User:   url.UserPassword("root", item.Group.Password),
	}

	// check if the API is available, we will make 120 connection attempts, the connection test will reply with "connection refused" while no os is available to respond, in that case we sleep for 10 seconds to give it some time to boot.
	/*
		err := retry(120, 1*time.Second, func() (err error) {
			test_ctx := context.Background()
			_, err = govmomi.NewClient(test_ctx, url, true)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Postconfig": "still attempting to connect to API",
				}).Debug(item.IP)
				// if we get "connection refused" we wait 10 seconds.
				match, _ := regexp.MatchString("refused", err.Error())
				if match {
					time.Sleep(10 * time.Second)
				}
			}
			return
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"postconfig": err,
			}).Info(item.IP)
		}
	*/

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
	timeout := 2
	for {
		if i > timeout {
			fmt.Println("timeout exceeded, failing")
			return
		}
		c, err = govmomi.NewClient(ctx, url, true)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
			i += 1
			fmt.Println("sleeping for 10 seconds")
			<-time.After(time.Second * 10)
			continue
		}
		break
	}

	// since we're always going to be talking directly to the host, dont asume connection through vCenter.
	host, err := find.NewFinder(c.Client).DefaultHostSystem(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Postconfig": err,
		}).Info(item.IP)
	}
	e, err := esxcli.NewExecutor(c.Client, host)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Postconfig": err,
		}).Info(item.IP)
	}

	if options.Domain {
		search := strings.Fields("network ip dns search add -d")
		search = append(search, item.Domain)
		_, err := e.Run(search)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"search domain": item.Domain,
		}).Info(item.IP)

		hd := string(item.Hostname + "." + item.Domain)
		fqdn := strings.Fields("system hostname set --fqdn")
		fqdn = append(fqdn, hd)
		_, err = e.Run(fqdn)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"fqdn": hd,
		}).Info(item.IP)
	}

	if options.NTP {
		//configure ntpd
		cmd := strings.Fields("system ntp set")
		for _, k := range strings.Split(item.Group.NTP, ",") {
			cmd = append(cmd, "--server", string(k))
		}
		_, err := e.Run(cmd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"ntpd": item.Group.NTP,
		}).Info(item.IP)
		//enable ntpd
		_, err = e.Run(strings.Fields("system ntp set --enabled true"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"ntpd": "Service enabled",
		}).Info(item.IP)

		s, err := host.ConfigManager().ServiceSystem(ctx)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}

		err = s.UpdatePolicy(ctx, "ntpd", string(types.HostServicePolicyOn))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"ntpd": "Startup Policy -> Start and stop with host",
		}).Info(item.IP)

		err = s.Start(ctx, "ntpd")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"ntpd": "Service started",
		}).Info(item.IP)
	}

	if options.Syslog {
		//configure Syslog
		cmd := strings.Fields("system syslog config set --loghost=" + item.Group.Syslog)
		spew.Dump(cmd)
		_, err := e.Run(cmd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}
		_, err = e.Run(strings.Fields("system syslog reload"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"syslog": "syslog configured",
		}).Info(item.IP)
	}

	if options.SSH {
		s, err := host.ConfigManager().ServiceSystem(ctx)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}

		err = s.UpdatePolicy(ctx, "TSM-SSH", string(types.HostServicePolicyOn))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"ssh": "Startup Policy -> Start and stop with host",
		}).Info(item.IP)

		err = s.Start(ctx, "TSM-SSH")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"ssh": "Service started",
		}).Info(item.IP)
	}

	if options.SuppressShellWarning {
		//Suppress any warnings that ESXi Console or SSH are enabled
		cmd := strings.Fields("system settings advanced set -o /UserVars/SuppressShellWarning -i 1")
		_, err := e.Run(cmd)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Postconfig": err,
			}).Info(item.IP)
		}
		logrus.WithFields(logrus.Fields{
			"ssh": "suppressing shell warnings",
		}).Info(item.IP)
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

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func CustomRetryTemporaryNetworkError(err error) (bool, time.Duration) {
	return vim25.IsTemporaryNetworkError(err), time.Second * 10
}
