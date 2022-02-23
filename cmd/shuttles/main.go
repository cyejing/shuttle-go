package main

import (
	"flag"
	config "github.com/cyejing/shuttle/core/config/server"
	"github.com/cyejing/shuttle/core/server"
	"github.com/cyejing/shuttle/pkg/logger"
)

var (
	configPath = flag.String("c", "", "config path")
)

func main() {
	flag.Parse()
	c, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}

	err = logger.InitLog(c.LogFile)
	if err != nil {
		panic(err)
	}

	server.Run(c)
}
