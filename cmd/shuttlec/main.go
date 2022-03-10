package main

import (
	"flag"
	"github.com/cyejing/shuttle/core/client"
	config "github.com/cyejing/shuttle/core/config/client"
	"github.com/cyejing/shuttle/pkg/logger"
	"github.com/google/gops/agent"
)

var (
	configPath = flag.String("c", "", "config path")
	debug      = flag.Bool("d", false, "debug mode")
)
var log = logger.NewLog()

func main() {
	flag.Parse()
	if *debug {
		log.Debug("open debug mode")
		if err := agent.Listen(agent.Options{}); err != nil {
			log.Fatal(err)
		}
	}

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
