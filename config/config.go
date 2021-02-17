package config

type Config struct {
	Debug   bool
	Port    int `default:"8080"`
	File    string
	Network Network
	SQLite  SQLte
}

type Network struct {
	Interfaces []string
}
