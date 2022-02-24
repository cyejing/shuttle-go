package server

import (
	"github.com/cyejing/shuttle/core/config"
	"github.com/cyejing/shuttle/pkg/logger"
	"github.com/cyejing/shuttle/pkg/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

var log = logger.NewLog()

//Config struct
type Config struct {
	Addrs     []Addr `yaml:"addrs"`
	LogFile   string `yaml:"logFile"`
	Gateway   Gateway
	Instances []Instance
	Trojan    Trojan
	Wormhole  Wormhole
}

type Addr struct {
	Addr string `yaml:"addr"`
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

//Gateway struct
type Gateway struct {
	Routes []Route
}

//Wormhole struct
type Wormhole struct {
	Passwords []string
}

//Trojan struct
type Trojan struct {
	Passwords []string
}

//Route struct
type Route struct {
	ID       string `yaml:"id"`
	Order    int
	Host     string
	Path     string
	Filters  []Filter
	Loggable bool
}

// GetFilter route get filter
func (r Route) GetFilter(name string) (Filter, error) {
	for _, filter := range r.Filters {
		if name == filter.Name {
			return filter, nil
		}
	}
	return *new(Filter), utils.NewErrf("not fount filter %s ", name)
}

// Instance struct
type Instance struct {
	Group        string
	URL          string `yaml:"url"`
	Weight       int
	RegisterTime int64
	Tags         []string
}

// Filter struct
type Filter struct {
	Name     string
	Params   interface{}
	Open     bool
	Loggable bool
}

// Password struct
type Password struct {
	Raw  string
	Hash string
}

// config var
var (
	defaultConfigPath = []string{"shuttles.yaml", "shuttles.yaml", "example/shuttles.yaml", "example/shuttles.yml"}
	GlobalConfig      = &Config{
		LogFile: "logs/shuttles.log",
	}
	Passwords   = make(map[string]*Password)
	WHPasswords = make(map[string]*Password)
)

//Load load config
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
			log.Infof("load config %v", config)
			break
		}
	default:
		data, err = ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("load config file %s err %v", path, err)
		}
		log.Infof("load config %s", path)
	}
	err = yaml.Unmarshal(data, GlobalConfig)
	initPasswords()
	return GlobalConfig, err
}

func initPasswords() {
	for _, raw := range GlobalConfig.Trojan.Passwords {
		hash := utils.SHA224String(config.TrojanSalt + raw + config.TrojanSalt)
		Passwords[hash] = &Password{
			Raw:  raw,
			Hash: hash,
		}
		hash2 := utils.SHA224String(raw)
		Passwords[hash2] = &Password{
			Raw:  raw,
			Hash: hash2,
		}
	}
	for _, raw := range GlobalConfig.Wormhole.Passwords {
		hash := utils.SHA224String(config.WormholeSalt + raw + config.WormholeSalt)
		WHPasswords[hash] = &Password{
			Raw:  raw,
			Hash: hash,
		}
	}
}
