//go:generate bash -c "go get github.com/swaggo/swag/cmd/swag && swag init"

package main

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/koding/multiconfig"
	"github.com/sirupsen/logrus"
	"honnef.co/go/tools/config"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var db *gorm.DB

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

	//check if database is present
	if _, err := os.Stat("sqlite-database.db"); os.IsNotExist(err) {
		//Database does not exist, so create it.

		file, err := os.Create("sqlite-database.db")
		if err != nil {
			logrus.Fatal(err.Error())
		}
		file.Close()
		logrus.Info("No database found, sqlite-database.db created")
	} else {
		//Database exists, moving on.

		logrus.Info("Existing database sqlite-database.db found")
	}

	//connect to sqlite databaes
	db, err = gorm.Open("sqlite3", "sqlite-database.db")
	if err != nil {
		logrus.Error("Failed to open the SQLite database.")
	}

	defer db.Close()

}
