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
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	if debug {
		c.Logger = logger.Default.LogMode(logger.Info)
	}

	//check if database is present
	if _, err := os.Stat("database/sqlite-database.db"); os.IsNotExist(err) {
		//Database does not exist, so create it.
		os.MkdirAll("database", os.ModePerm)
		logrus.Info("No database found, creating database/sqlite-database.db")
		file, err := os.Create("database/sqlite-database.db")
		if err != nil {
			logrus.Fatal(err.Error())
		}
		file.Close()
		logrus.Info("database/sqlite-database.db created")
	} else {
		//Database exists, moving on.
		logrus.Info("Existing database sqlite-database.db found")
	}

	var err error

	DB, err = gorm.Open(sqlite.Open("database/sqlite-database.db"), c)
	if err != nil {
		logrus.Error("Failed to open the SQLite database.")
	}
}
