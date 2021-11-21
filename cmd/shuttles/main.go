package main

import (
	"flag"
	"github.com/cyejing/shuttle/core/server"
	config "github.com/cyejing/shuttle/pkg/config/server"
	"github.com/cyejing/shuttle/pkg/log"
)

var (
	configPath = flag.String("c", "", "config path")
)

func main() {
	flag.Parse()
	if _, err := config.Load(*configPath); err != nil {
		panic(err)
	}
	log.Debugf("load config %v", config.GetConfig())
	server.StartWebServer()
}
