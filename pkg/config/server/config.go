package server

import (
	logger "github.com/cyejing/shuttle/pkg/logger"
	"github.com/cyejing/shuttle/pkg/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

var log = logger.NewLog()

type Config struct {
	Addr      string `yaml:"addr"`
	SslAddr   string `yaml:"sslAddr"`
	Cert      string
	Key       string
	Passwords []string
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

func (r Route) GetFilter(name string) Filter {
	for _, filter := range r.Filters {
		if name == filter.Name {
			return filter
		}
	}
	return *new(Filter)
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

type Password struct {
	Raw  string
	Hash string
}

var (
	defaultConfigPath = []string{"shuttles.yaml", "shuttles.yaml", "example/shuttles.yaml", "example/shuttles.yml"}
	GlobalConfig      = &Config{
		Addr:    "127.0.0.1:4880",
		SslAddr: "127.0.0.1:4843",
	}
	Passwords = make(map[string]*Password)
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
			log.Infof("load config %s", config)
			break
		}
	default:
		log.Infof("load config %s", path)
		data, err = ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("load config file %s err %v", path, err)
		}
	}
	err = yaml.Unmarshal(data, GlobalConfig)
	initPasswords()
	return GlobalConfig, err
}

func GetConfig() *Config {
	return GlobalConfig
}

func initPasswords() {
	for _, raw := range GlobalConfig.Passwords {
		hash := utils.SHA224String(raw)
		Passwords[hash] = &Password{
			Raw:  raw,
			Hash: hash,
		}
	}
}
