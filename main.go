//go:generate bash -c "go get github.com/swaggo/swag/cmd/swag && swag init"
//go:generate bash -c "cd web && rm -rf ./web/dist && npm install && npm run build && cd .. && go get github.com/rakyll/statik && statik -src ./web/dist -f"

package main

import (
	"flag"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/maxiepax/go-via/api"
	"github.com/maxiepax/go-via/config"
	ca "github.com/maxiepax/go-via/crypto"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/maxiepax/go-via/secrets"
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

	//try to load environment variables and flags.
	err := d.Load(conf)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Info("failed to load config")
	}

	//if a file has been implied, also load the content of the configuration file.
	if conf.File != "" {
		d = multiconfig.NewWithPath(conf.File)

		err = d.Load(conf)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Info("failed to load config")
		}
	}

	//validate configuration file
	err = d.Validate(conf)
	if err != nil {
		flag.Usage()
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Info("failed to load config")
	}

	//if no environemnt variables, or configuration file has been declared, serve on all interfaces.
	if len(conf.Network.Interfaces) == 0 {
		logrus.Warning("no interfaces have been configured, trying to find interfaces to serve to, will serve on all.")
		i, err := net.Interfaces()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Info("failed to find a usable interface")
		}
		for _, v := range i {
			// dont use loopback interfaces
			if v.Flags&net.FlagLoopback != 0 {
				continue
			}
			// dont use ptp interfaces
			if v.Flags&net.FlagPointToPoint != 0 {
				continue
			}
			_, _, err := findIPv4Addr(&v)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"err":   err,
					"iface": v.Name,
				}).Warning("interaces does not have a usable ipv4 address")
				continue
			}
			conf.Network.Interfaces = append(conf.Network.Interfaces, v.Name)
		}
	}

	// load secrets key
	key := secrets.Init()

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
	err = db.DB.AutoMigrate(&models.Pool{}, &models.Address{}, &models.Option{}, &models.DeviceClass{}, &models.Group{}, &models.Image{}, &models.User{})
	if err != nil {
		logrus.Fatal(err)
	}

	//create the device classes for x86 and arm
	var vc models.DeviceClass
	//64bit x86 UEFI
	if res := db.DB.FirstOrCreate(&vc, models.DeviceClass{DeviceClassForm: models.DeviceClassForm{Name: "PXE-UEFI_x64", VendorClass: "PXEClient:Arch:00007"}}); res.Error != nil {
		logrus.Warning(res.Error)
	}
	//64bit ARM UEFI
	if res := db.DB.FirstOrCreate(&vc, models.DeviceClass{DeviceClassForm: models.DeviceClassForm{Name: "PXE-UEFI_ARM64", VendorClass: "PXEClient:Arch:00011"}}); res.Error != nil {
		logrus.Warning(res.Error)
	}

	//create admin user if it doesn't exist
	var adm models.User
	hp := api.HashAndSalt([]byte("VMware1!"))
	if res := db.DB.Where(models.User{UserForm: models.UserForm{Username: "admin"}}).Attrs(models.User{UserForm: models.UserForm{Password: hp}}).FirstOrCreate(&adm); res.Error != nil {
		logrus.Warning(res.Error)
	}

	// DHCPd
	for _, v := range conf.Network.Interfaces {
		go serve(v)
	}

	// TFTPd
	go TFTPd(conf)

	//REST API
	r := gin.New()
	r.Use(cors.Default())

	statikFS, err := fs.New()
	if err != nil {
		logrus.Fatal(err)
	}

	// ks.cfg is served at top to not place it behind BasicAuth
	r.GET("ks.cfg", api.Ks(key))

	// middleware to check if user is logged in
	r.Use(func(c *gin.Context) {
		username, password, hasAuth := c.Request.BasicAuth()
		if !hasAuth {
			logrus.WithFields(logrus.Fields{
				"login": "unauthorized request",
			}).Info("auth")
			c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//get the user that is trying to authenticate
		var user models.User
		if res := db.DB.Select("username", "password").Where("username = ?", username).First(&user); res.Error != nil {
			logrus.WithFields(logrus.Fields{
				"username": username,
				"status":   "supplied username does not exist",
			}).Info("auth")
			c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//check if passwords match
		if api.ComparePasswords(user.Password, []byte(password), username) {
			logrus.WithFields(logrus.Fields{
				"username": username,
				"status":   "successfully authenticated",
			}).Debug("auth")
		} else {
			logrus.WithFields(logrus.Fields{
				"username": username,
				"status":   "invalid password supplied",
			}).Info("auth")
			c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	})

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
			groups.POST("", api.CreateGroup(key))
			groups.PATCH(":id", api.UpdateGroup(key))
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

		users := v1.Group("/users")
		{
			users.GET("", api.ListUsers)
			users.GET(":id", api.GetUser)
			users.POST("", api.CreateUser)
			users.PATCH(":id", api.UpdateUser)
			users.DELETE(":id", api.DeleteUser)
		}

		postconfig := v1.Group("/postconfig")
		{
			postconfig.GET("", api.PostConfig(key))
			postconfig.GET(":id", api.PostConfigID(key))
		}

		v1.GET("log", logServer.Handle)

		v1.GET("version", api.Version(version, commit, date))
	}

	/*	r.GET("postconfig", api.PostConfig) */

	// check if ./cert/server.crt exists, if not we will create the folder, and initiate a new CA and a self-signed certificate
	crt, err := os.Stat("./cert/server.crt")
	if os.IsNotExist(err) {
		// create folder for certificates
		logrus.WithFields(logrus.Fields{
			"certificate": "server.crt does not exist, initiating new CA and creating self-signed ceritificate server.crt",
		}).Info("cert")
		os.MkdirAll("cert", os.ModePerm)
		ca.CreateCA()
		ca.CreateCert("./cert", "server", "server")
	} else {
		logrus.WithFields(logrus.Fields{
			crt.Name(): "server.crt found",
		}).Info("cert")
	}
	//enable HTTPS
	listen := ":" + strconv.Itoa(conf.Port)
	logrus.WithFields(logrus.Fields{
		"port": listen,
	}).Info("Webserver")
	err = r.RunTLS(listen, "./cert/server.crt", "./cert/server.key")

	logrus.WithFields(logrus.Fields{
		"error": err,
	}).Error("Webserver")

}
