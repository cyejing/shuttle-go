package main

import (
	"flag"
	"github.com/cyejing/shuttle/core/client"
	config "github.com/cyejing/shuttle/core/config/client"
	"github.com/cyejing/shuttle/pkg/logger"
)

var (
	configPath = flag.String("c", "", "config path")
)
var log = logger.NewLog()

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

	client.Run(c)
}
