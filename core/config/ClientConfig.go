package config

import (
	"github.com/cyejing/shuttle/pkg/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)


//ClientConfig struct
type ClientConfig struct {
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
	defaultClientConfigPath = []string{
		"shuttlec-socks.yaml",
		"shuttlec-wormhole.yaml",
		//"example/shuttlec-socks.yaml",
		"example/shuttlec-wormhole.yaml",
	}
	GlobalClientConfig = &ClientConfig{
		SockAddr:  "127.0.0.1:1080",
		LogFile:   "logs/shuttlec.log",
		SSLEnable: true,
	}
)

//LoadClient load config
func LoadClient(path string) (config ClientConfig, err error) {
	var data []byte
	switch path {
	case "":
		for _, config := range defaultClientConfigPath {
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
	err = yaml.Unmarshal(data, GlobalClientConfig)
	return *GlobalClientConfig, err
}


func (c *ClientConfig) IsSocks() bool {
	return "socks" == c.RunType
}

func (c *ClientConfig) IsWormhole() bool {
	return "wormhole" == c.RunType
}

func (c *ClientConfig) GetHash() string {
	if c.IsSocks() {
		return utils.SHA224String(TrojanSalt + c.Password + TrojanSalt)
	} else if c.IsWormhole() {
		return utils.SHA224String(WormholeSalt + c.Password + WormholeSalt)
	} else {
		log.Errorf("unknown run type %s,please check config", c.RunType)
	}
	return ""
}

