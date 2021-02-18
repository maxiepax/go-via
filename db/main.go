package db

import (
	"os"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(debug bool) {

	c := &gorm.Config{
		SkipDefaultTransaction: true,
	}

	if debug {
		c.Logger = logger.Default.LogMode(logger.Info)
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

	var err error

	DB, err = gorm.Open(sqlite.Open("sqlite-database.db"), c)
	if err != nil {
		logrus.Error("Failed to open the SQLite database.")
	}
}
