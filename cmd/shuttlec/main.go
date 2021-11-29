package main

import (
	"flag"
	"github.com/cyejing/shuttle/core/server"
	"github.com/cyejing/shuttle/pkg/codec"
	config "github.com/cyejing/shuttle/pkg/config/client"
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

	socks5 := &server.Socks5Server{
		DialFunc: codec.DialTrojan,
	}
	panic(socks5.ListenAndServe("tcp", c.LocalAddr))
}
