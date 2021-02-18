//go:generate bash -c "go get github.com/swaggo/swag/cmd/swag && swag init"

package main

import (
	"github.com/koding/multiconfig"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/sirupsen/logrus"
	"honnef.co/go/tools/config"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// @title go-via
// @version 0.1
// @description VMware Imaging Appliances written in GO with full HTTP-REST

// @BasePath /v1

func main() {

	//setup logging
	logrus.WithFields(logrus.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Infof("Startup")

	//enable config
	d := multiconfig.New()

	conf := new(config.Config)

	err := d.Load(conf)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Fatalf("failed to load config")
	}

	//connect to database
	db.Connect(true)

	//migrate all models
	err = db.DB.AutoMigrate(&models.Pool{}, &models.Address{}, &models.Option{}, &models.DeviceClass{})
	if err != nil {
		logrus.Fatal(err)
	}

}
