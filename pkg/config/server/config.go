package server

import (
	"github.com/cyejing/shuttle/pkg/log"
	"github.com/cyejing/shuttle/pkg/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	Addr      string `yaml:"addr"`
	Ssl       *Ssl
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

type Ssl struct {
	Cert string
	Key  string
}

type Password struct {
	raw  string
	hash string
}

var (
	defaultConfigPath = []string{"shuttles.yaml", "shuttles.yaml", "example/shuttles.yaml", "example/shuttles.yml"}
	globalConfig      = &Config{
		Addr: "127.0.0.1:4843",
		Ssl:  &Ssl{},
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
			log.L.Debugf("load config %s", config)
			break
		}
	default:
		log.L.Infof("load config %s", path)
		data, err = ioutil.ReadFile(path)
		if err != nil {
			log.L.Fatal("load config file %s err", path, err)
		}
	}
	err = yaml.Unmarshal(data, globalConfig)
	initPasswords()
	return globalConfig, err
}

func GetConfig() *Config {
	return globalConfig
}

func initPasswords() {
	for _, raw := range globalConfig.Passwords {
		hash := utils.SHA224String(raw)
		Passwords[hash] = &Password{
			raw:  raw,
			hash: hash,
		}
	}
}
