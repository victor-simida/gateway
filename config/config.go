package config

import (
	"gateway/Godeps/_workspace/src/github.com/jinzhu/configor"
	"fmt"
)

type Config struct {
	HttpsAddr      []Address    `yaml:"https_addr" form:"HttpsAddr"`
	HttpAddr       []Address    `yaml:"http_addr" form:"HttpAddr"`

	HttpBatchAddr  []Address  `yaml:"http_batch_addr" form:"http_batch_addr"`
	HttpsBatchAddr []Address  `yaml:"https_batch_addr" form:"https_batch_addr"`
	ServerPort     string       `yaml:"server_port" form:"ServerPort"`
}

type Address struct {
	Suffix string    `yaml:"suffix" form:"suffix"`
	Prefix string     `yaml:"prefix" form:"suffix"`
	Host   string  `yaml:"host" form:"host"`
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
