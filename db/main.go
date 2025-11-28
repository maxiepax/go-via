package db

import (
	"fmt"
	"os"
	"regexp"

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
		err = os.MkdirAll("database", os.ModePerm)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Warn("could not create database directory")
		}
		logrus.Info("No database found, creating database/sqlite-database.db")
		file, err := os.Create("database/sqlite-database.db")
		if err != nil {
			logrus.Fatal(err.Error())
		}
		err = file.Close()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Warn("could not close database/sqlite-database.db")
		}
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

func Migrate(models []interface{}) {
	migrateErr := DB.AutoMigrate(models[0])
	if migrateErr != nil {
		index, err := getIndexFromErr(migrateErr)
		if err != nil {
			fmt.Println(err)
		}

		if indexExists(index) {
			fmt.Printf("index %s exists, dropping", index)
			err := DB.Migrator().DropIndex(models[0], index)
			if err != nil {
				fmt.Println(err)
			}
			Migrate(models)

		} else {
			fmt.Printf("index %s does not exist", index)

		}
	}
	if len(models) > 1 {
		Migrate(models[1:])
	}

}

func indexExists(index string) bool {
	rows, err := DB.Raw("SELECT name, tbl_name FROM sqlite_master WHERE type = 'index'").Rows()
	if err != nil {
		fmt.Println("Error fetching indexes:", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Println("Error closing rows:", err)
		}
	}()

	fmt.Println("Existing indexes:")
	for rows.Next() {
		var indexName, tableName string
		err := rows.Scan(&indexName, &tableName)
		if err != nil {
			fmt.Println("Error scanning rows:", err)
			continue
		}
		fmt.Printf("Index: %s, Table: %s\n", indexName, tableName)

		if indexName == index {
			return true
		}
	}
	return false
}

func getIndexFromErr(err error) (string, error) {
	errorMessage := err.Error()
	// Regular expression to extract the variable
	re := regexp.MustCompile(`index (\w+) already exists`)

	// Find the match
	match := re.FindStringSubmatch(errorMessage)
	if len(match) > 1 {
		variable := match[1]
		fmt.Println("Extracted variable:", variable)
		return variable, nil
	} else {
		fmt.Println("No variable found in the error message")
		return "", fmt.Errorf("no variable found in the error message %s", errorMessage)
	}
}
