//go:generate bash -c "go get github.com/swaggo/swag/cmd/swag && swag init"

package main

import (
	"strconv"

	"github.com/maxiepax/go-via/api"
	"github.com/maxiepax/go-via/config"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/koding/multiconfig"

	"github.com/sirupsen/logrus"
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
	err = db.DB.AutoMigrate(&models.Pool{}, &models.Address{}, &models.Option{}, &models.DeviceClass{}, &models.Group{}, &models.Image{})
	if err != nil {
		logrus.Fatal(err)
	}

	//REST API

	r := gin.New()
	r.Use(cors.Default())

	v1 := r.Group("/v1")
	{
		//v1.GET("log", logServer.Handle)

		pools := v1.Group("/pools")
		{
			pools.GET("", api.ListPools)
			pools.GET(":id", api.GetPool)
			pools.POST("/search", api.SearchPool)
			pools.POST("", api.CreatePool)
			pools.PATCH(":id", api.UpdatePool)
			pools.DELETE(":id", api.DeletePool)

			pools.GET(":id/next", api.GetNextFreeIP)
		}
		relay := v1.Group("/relay")
		{
			relay.GET(":relay", api.GetPoolByRelay)
		}

		addresses := v1.Group("/addresses")
		{
			addresses.GET("", api.ListAddresses)
			addresses.GET(":id", api.GetAddress)
			addresses.POST("/search", api.SearchAddress)
			addresses.POST("", api.CreateAddress)
			addresses.PATCH(":id", api.UpdateAddress)
			addresses.DELETE(":id", api.DeleteAddress)
		}

		options := v1.Group("/options")
		{
			options.GET("", api.ListOptions)
			options.GET(":id", api.GetOption)
			options.POST("/search", api.SearchOption)
			options.POST("", api.CreateOption)
			options.PATCH(":id", api.UpdateOption)
			options.DELETE(":id", api.DeleteOption)
		}

		device_class := v1.Group("/device_classes")
		{
			device_class.GET("", api.ListDeviceClasses)
			device_class.GET(":id", api.GetDeviceClass)
			device_class.POST("/search", api.SearchDeviceClass)
			device_class.POST("", api.CreateDeviceClass)
			device_class.PATCH(":id", api.UpdateDeviceClass)
			device_class.DELETE(":id", api.DeleteDeviceClass)
		}
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	listen := ":" + strconv.Itoa(conf.Port)
	logrus.WithFields(logrus.Fields{
		"address": listen,
	}).Info("Webserver")
	err = r.Run(listen)

	logrus.WithFields(logrus.Fields{
		"error": err,
	}).Error("Webserver")

}
