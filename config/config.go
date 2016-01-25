package config

import (
    "gateway/Godeps/_workspace/src/github.com/jinzhu/configor"
    "fmt"
)

type Config struct {
	HttpsAddr []Address    `yaml:"https_addr" form:"HttpsAddr"`
	HttpAddr  []Address    `yaml:"http_addr" form:"HttpAddr"`
	ServerPort string	   `yaml:"server_port" form:"ServerPort"`
}

type Address struct {
	Suffix string    `yaml:"suffix" form:"Suffix"`
	Prefix string     `yaml:"prefix" form:"Prefix"`
	Method string    `yaml:"method" form:"Method"`
}

var Settings *Config

func init() {
    var path string
	Settings = &Config{}
    path = "./config/config.yml"

    if err := configor.Load(Settings, path); err != nil {
        panic(err)
    }
    fmt.Println("loading settings")

}
