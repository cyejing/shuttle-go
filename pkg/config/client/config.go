package client

import (
	"github.com/cyejing/shuttle/pkg/config"
	"github.com/cyejing/shuttle/pkg/logger"
	"github.com/cyejing/shuttle/pkg/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

var log = logger.NewLog()

//Config struct
type Config struct {
	RunType    string `yaml:"runType"`
	Name       string `yaml:"name"`
	SockAddr   string `yaml:"sockAddr"`
	RemoteAddr string `yaml:"remoteAddr"`
	SSLEnable  bool   `yaml:"sslEnable"`
	Password   string `yaml:"password"`
	LogFile    string `yaml:"logFile"`

	Ships []Ship
}

type Ship struct {
	Name string
	RemoteAddr string `yaml:"remoteAddr"`
	LocalAddr string `yaml:"localAddr"`
}

//global config
var (
	defaultConfigPath = []string{
		//"shuttlec-socks.yaml",
		//"shuttlec-wormhole.yaml",
		//"example/shuttlec-socks.yaml",
		"example/shuttlec-wormhole.yaml",
	}
	GlobalConfig = &Config{
		SockAddr:  "127.0.0.1:1080",
		LogFile:   "logs/shuttlec.log",
		SSLEnable: true,
	}
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
		log.Infof("load config %s", path)
		if err != nil {
			log.Fatalf("load config file %s err %v", path, err)
		}
		data, err = ioutil.ReadFile(path)
	}
	err = yaml.Unmarshal(data, GlobalConfig)
	return GlobalConfig, err
}


func (c *Config) IsSocks() bool {
	return "socks" == c.RunType
}

func (c *Config) IsWormhole() bool {
	return "wormhole" == c.RunType
}

func (c *Config) GetHash() string {
	if c.IsSocks() {
		return utils.SHA224String(config.TrojanSalt + c.Password + config.TrojanSalt)
	} else if c.IsWormhole() {
		return utils.SHA224String(config.WormholeSalt + c.Password + config.WormholeSalt)
	} else {
		log.Errorf("unknown run type %s,please check config", c.RunType)
	}
	return ""
}

