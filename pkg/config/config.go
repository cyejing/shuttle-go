package config

import (
	"github.com/cyejing/shuttle/pkg/log"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	Addr      string `yaml:"addr"`
	Ssl       Ssl
	Routes    []Route
	Instances []Instance
}

type Route struct {
	Id       string
	Order    int
	Host     string
	Path     string
	Filters  []Filter
	Loggable bool
}

type Instance struct {
	Group        string
	Url          string
	Weight       int
	RegisterTime int64
	Tags         []string
}

type Filter struct {
	Name     string
	Params   interface{}
	Open     bool
	Loggable bool
}

type Ssl struct {
	Enable bool
	Cert   string
	Key    string
}

var (
	defaultConfigPath = []string{"shuttles.yaml", "shuttles.yaml", "config/shuttles.yaml", "config/shuttles.yml"}
	globalConfig      = &Config{
		Addr: "127.0.0.1:4843",
		Ssl:  Ssl{Enable: true},
	}
)

func Load(path string) (config *Config, err error) {
	var data []byte
	switch path {
	case "":
		for _, config := range defaultConfigPath {
			data, err = ioutil.ReadFile(config)
			if err != nil {
				// is ok
				continue
			}
			break
		}
	default:
		data, err = ioutil.ReadFile(path)
		if err != nil {
			log.Fatal("load config file %s err", path, err)
		}
	}
	err = yaml.Unmarshal(data, globalConfig)
	return globalConfig, err
}

func GetConfig() *Config {
	return globalConfig
}
