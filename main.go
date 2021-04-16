//go:generate bash -c "go get github.com/swaggo/swag/cmd/swag && swag init"
//go:generate bash -c "cd web && rm -rf ./web/dist && npm install && npm run build && cd .. && go get github.com/rakyll/statik && statik -src ./web/dist -f"

package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/maxiepax/go-via/api"
	"github.com/maxiepax/go-via/config"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/maxiepax/go-via/websockets"
	"github.com/rakyll/statik/fs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/koding/multiconfig"

	"github.com/sirupsen/logrus"

	_ "github.com/maxiepax/go-via/docs"
	_ "github.com/maxiepax/go-via/statik"
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

	logServer := websockets.NewLogServer()
	logrus.AddHook(logServer.Hook)

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

	if conf.File != "" {
		d = multiconfig.NewWithPath(conf.File)

		err = d.Load(conf)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Fatalf("failed to load config")
		}
	}

	err = d.Validate(conf)
	if err != nil {
		flag.Usage()
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Fatalf("failed to load config")
	}

	//connect to database
	//db.Connect(true)
	if conf.Debug {
		db.Connect(true)
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		db.Connect(false)
		gin.SetMode(gin.ReleaseMode)
	}

	//migrate all models
	err = db.DB.AutoMigrate(&models.Pool{}, &models.Address{}, &models.Option{}, &models.DeviceClass{}, &models.Group{}, &models.Image{})
	if err != nil {
		logrus.Fatal(err)
	}

	var vc models.DeviceClass

	if res := db.DB.FirstOrCreate(&vc, models.DeviceClass{DeviceClassForm: models.DeviceClassForm{Name: "PXE-UEFI_x86", VendorClass: "PXEClient:Arch:00007"}}); res.Error != nil {
		logrus.Warning(res.Error)
	}

	// DHCPd

	for _, v := range conf.Network.Interfaces {
		go serve(v)
	}

	// TFTPd

	go TFTPd()

	//REST API

	r := gin.New()
	r.Use(cors.Default())

	statikFS, err := fs.New()
	if err != nil {
		logrus.Fatal(err)
	}
	r.NoRoute(func(c *gin.Context) {
		c.Request.URL.Path = "/web/" // force us to always return index.html and not the requested page to be compatible with HTML5 routing
		http.FileServer(statikFS).ServeHTTP(c.Writer, c.Request)
	})
	ui := r.Group("/")
	{
		ui.GET("/web/*all", gin.WrapH(http.FileServer(statikFS)))

		ui.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

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

		deviceClass := v1.Group("/device_classes")
		{
			deviceClass.GET("", api.ListDeviceClasses)
			deviceClass.GET(":id", api.GetDeviceClass)
			deviceClass.POST("/search", api.SearchDeviceClass)
			deviceClass.POST("", api.CreateDeviceClass)
			deviceClass.PATCH(":id", api.UpdateDeviceClass)
			deviceClass.DELETE(":id", api.DeleteDeviceClass)
		}

		groups := v1.Group("/groups")
		{
			groups.GET("", api.ListGroups)
			groups.GET(":id", api.GetGroup)
			groups.POST("", api.CreateGroup)
			groups.PATCH(":id", api.UpdateGroup)
			groups.DELETE(":id", api.DeleteGroup)
		}

		images := v1.Group("/images")
		{
			images.GET("", api.ListImages)
			images.GET(":id", api.GetImage)
			images.POST("", api.CreateImage(conf))
			images.PATCH(":id", api.UpdateImage)
			images.DELETE(":id", api.DeleteImage)
		}

		v1.GET("log", logServer.Handle)
	}

	r.GET("ks.cfg", api.Ks)

	r.GET("postconfig", api.PostConfig)

	listen := ":" + strconv.Itoa(conf.Port)
	logrus.WithFields(logrus.Fields{
		"address": listen,
	}).Info("Webserver")
	err = r.Run(listen)

	logrus.WithFields(logrus.Fields{
		"error": err,
	}).Error("Webserver")

}
