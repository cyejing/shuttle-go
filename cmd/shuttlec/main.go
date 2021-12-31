package main

import (
	"flag"
	"github.com/cyejing/shuttle/core/client"
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

	switch c.RunType {
	case "socks":
		socks5 := &client.Socks5Server{
			Config:   c,
			DialFunc: codec.DialTrojan,
		}
		panic(socks5.ListenAndServe("tcp", c.LocalAddr))
	case "wormhole":
		wormhole := &client.Wormhole{
			Config: c,
			Name:   c.Name,
		}
		panic(wormhole.DialRemote("tcp", c.RemoteAddr))
	}

}
