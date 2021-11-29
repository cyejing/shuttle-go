package client

import (
	"github.com/cyejing/shuttle/pkg/logger"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

var log = logger.NewLog()

//Config struct
type Config struct {
	RunType    string `yaml:"runType"`
	LocalAddr  string `yaml:"localAddr"`
	RemoteAddr string `yaml:"remoteAddr"`
	Password   string `yaml:"password"`
	LogFile    string `yaml:"logFile"`
}

//global config
var (
	defaultConfigPath = []string{"shuttlec.yaml", "shuttlec.yaml", "example/shuttlec.yaml", "example/shuttlec.yml"}
	GlobalConfig      = &Config{
		LocalAddr: "127.0.0.1:1080",
		LogFile:   "logs/shuttlec.log",
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
			log.Infof("load config %s", config)
			break
		}
	default:
		log.Infof("load config %s", config)
		if err != nil {
			log.Fatalf("load config file %s err %v", path, err)
		}
		data, err = ioutil.ReadFile(path)
	}
	err = yaml.Unmarshal(data, GlobalConfig)
	return GlobalConfig, err
}

//GetConfig get config
func GetConfig() *Config {
	return GlobalConfig
}
