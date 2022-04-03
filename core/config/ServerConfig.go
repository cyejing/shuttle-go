package config

import (
	"github.com/cyejing/shuttle/pkg/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

//ServerConfig struct
type ServerConfig struct {
	Addrs     []Addr `yaml:"addrs"`
	LogFile   string `yaml:"logFile"`
	Instances []Instance
	Trojan    Trojan
	Wormhole  Wormhole
}

type Addr struct {
	Addr string `yaml:"addr"`
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

//Wormhole struct
type Wormhole struct {
	Passwords   []string
	PasswordMap map[string]*Password
}

//Trojan struct
type Trojan struct {
	Addr        string `yaml:"addr"`
	Passwords   []string
	PasswordMap map[string]*Password
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
	defaultServerConfigPath = []string{"shuttles.yaml", "shuttles.yaml", "example/shuttles.yaml", "example/shuttles.yml"}
	GlobalServerConfig      = &ServerConfig{
		LogFile: "logs/shuttles.log",
	}
	//TrojanPasswords   = make(map[string]*Password)
	//WormholePasswords = make(map[string]*Password)
)

//LoadServer load config
func LoadServer(path string) (config ServerConfig, err error) {
	var data []byte
	switch path {
	case "":
		for _, config := range defaultServerConfigPath {
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
	err = yaml.Unmarshal(data, GlobalServerConfig)
	initPasswords()
	return *GlobalServerConfig, err
}

func initPasswords() {
	for _, raw := range GlobalServerConfig.Trojan.Passwords {
		hash := utils.SHA224String(TrojanSalt + raw + TrojanSalt)
		GlobalServerConfig.Trojan.PasswordMap[hash] = &Password{
			Raw:  raw,
			Hash: hash,
		}
		hash2 := utils.SHA224String(raw)
		GlobalServerConfig.Trojan.PasswordMap[hash2] = &Password{
			Raw:  raw,
			Hash: hash2,
		}
	}
	for _, raw := range GlobalServerConfig.Wormhole.Passwords {
		hash := utils.SHA224String(WormholeSalt + raw + WormholeSalt)
		GlobalServerConfig.Wormhole.PasswordMap[hash] = &Password{
			Raw:  raw,
			Hash: hash,
		}
	}
}
