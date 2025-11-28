package config

import (
	"flag"
	"net"

	"github.com/koding/multiconfig"
	"github.com/maxiepax/go-via/dhcpd"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Debug       bool
	Port        int `default:"8443"`
	File        string
	Network     Network
	DisableDhcp bool
}

type Network struct {
	Interfaces []string
}

var conf *Config

func Get() *Config {
	return conf
}

func Set(c *Config) {
	conf = c
}

func Load() *Config {
	d := multiconfig.New()

	c := new(Config)

	err := d.Load(c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Fatalf("failed to load config")
	}

	if c.File != "" {
		d = multiconfig.NewWithPath(c.File)

		err = d.Load(c)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Fatalf("failed to load config")
		}
	}

	err = d.Validate(c)
	if err != nil {
		flag.Usage()
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Fatalf("failed to load config")
	}

	if len(c.Network.Interfaces) == 0 {
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
			_, _, err := dhcpd.FindIPv4Addr(&v)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"err":   err,
					"iface": v.Name,
				}).Warning("interface does not have a usable ipv4 address")
				continue
			}
			c.Network.Interfaces = append(c.Network.Interfaces, v.Name)
		}
	}

	conf = c

	return c
}
