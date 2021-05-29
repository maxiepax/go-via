package config

type Config struct {
	Debug   bool
	Port    int `default:"8443"`
	File    string
	Network Network
}

type Network struct {
	Interfaces []string //`default:"all"`
}
