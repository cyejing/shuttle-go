package main

import (
	"flag"
	"github.com/cyejing/shuttle/core/config"
	"github.com/cyejing/shuttle/core/server"
	"github.com/cyejing/shuttle/pkg/logger"
	"github.com/cyejing/shuttle/pkg/utils"
)

var (
	configPath = flag.String("c", "", "config path")
	debug      = flag.Bool("d", false, "debug mode")
)

func main() {
	flag.Parse()
	c, err := config.LoadServer(*configPath)
	if err != nil {
		panic(err)
	}

	err = logger.InitLog(c.LogFile)
	if err != nil {
		panic(err)
	}

	if *debug {
		utils.OpenAgent()
	}

	server.Run(c)
}
