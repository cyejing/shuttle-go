package client

import (
	"github.com/cyejing/shuttle/pkg/logger"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

var log = logger.NewLog()

type Config struct {
	RunType    string `yaml:"runType"`
	LocalAddr  string `yaml:"localAddr"`
	RemoteAddr string `yaml:"remoteAddr"`
	Password   string
}

var (
	defaultConfigPath = []string{"shuttlec.yaml", "shuttlec.yaml", "example/shuttlec.yaml", "example/shuttlec.yml"}
	globalConfig      = &Config{
		LocalAddr: "127.0.0.1:1080",
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
			log.Fatalf("load config file %s err %v", path, err)
		}
	}
	err = yaml.Unmarshal(data, globalConfig)
	return globalConfig, err
}

func GetConfig() *Config {
	return globalConfig
}
