package api

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/govc/host/esxcli"
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

func ProvisioningWorker(item models.Address) {

	//create empty model and load it with the json content from database
	options := models.GroupOptions{}
	json.Unmarshal(item.Group.Options, &options)
	spew.Dump(options)
	logrus.WithFields(logrus.Fields{
		"Started worker for ": item.Hostname,
	}).Debug("host")

	// connection info
	url := &url.URL{
		Scheme: "https",
		Host:   "192.168.1.175", //todo
		Path:   "sdk",
		User:   url.UserPassword("root", "VMware1!"), //todo
	}

	ctx := context.Background()
	c, err := govmomi.NewClient(ctx, url, true)
	if err != nil {
		log.Fatal(err)
	}

	// since we're always going to be talking directly to the host, dont asume connection through vCenter.
	host, err := find.NewFinder(c.Client).DefaultHostSystem(ctx)
	if err != nil {
		log.Fatal(err)
	}
	e, err := esxcli.NewExecutor(c.Client, host)
	if err != nil {
		log.Fatal(err)
	}

	if options.Domain {
		cmd := strings.Fields("network ip dns search add -d")
		cmd = append(cmd, item.Domain)
		_, err := e.Run(cmd)
		if err != nil {
			log.Fatal(err)
		}
		logrus.WithFields(logrus.Fields{
			"domain": item.Domain,
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
			log.Fatal(err)
		}
		logrus.WithFields(logrus.Fields{
			"ntpd": item.Group.NTP,
		}).Info(item.IP)
		//enable ntpd
		_, err = e.Run(strings.Fields("system ntp set --enabled true"))
		if err != nil {
			log.Fatal(err)
		}
		logrus.WithFields(logrus.Fields{
			"ntpd": "Service enabled",
		}).Info(item.IP)

		s, err := host.ConfigManager().ServiceSystem(ctx)
		if err != nil {
			log.Fatal(err)
		}

		err = s.UpdatePolicy(ctx, "ntpd", string(types.HostServicePolicyOn))
		if err != nil {
			log.Fatal(err)
		}
		logrus.WithFields(logrus.Fields{
			"ntpd": "Startup Policy -> Start and stop with host",
		}).Info(item.IP)

		err = s.Start(ctx, "ntpd")
		if err != nil {
			log.Fatal(err)
		}
		logrus.WithFields(logrus.Fields{
			"ntpd": "Service started",
		}).Info(item.IP)
	}

	if options.SSH {
		s, err := host.ConfigManager().ServiceSystem(ctx)
		if err != nil {
			log.Fatal(err)
		}

		err = s.UpdatePolicy(ctx, "TSM-SSH", string(types.HostServicePolicyOn))
		if err != nil {
			log.Fatal(err)
		}
		logrus.WithFields(logrus.Fields{
			"ssh": "Startup Policy -> Start and stop with host",
		}).Info(item.IP)

		err = s.Start(ctx, "TSM-SSH")
		if err != nil {
			log.Fatal(err)
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
			log.Fatal(err)
		}
		logrus.WithFields(logrus.Fields{
			"ssh": "suppressing shell warnings",
		}).Info(item.IP)
	}
}
