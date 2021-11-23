package main

import (
	"flag"
	"github.com/cyejing/shuttle/core/server"
	"github.com/cyejing/shuttle/pkg/codec"
	config "github.com/cyejing/shuttle/pkg/config/client"
)

var (
	configPath = flag.String("c", "", "config path")
)

func main() {
	flag.Parse()
	if _, err := config.Load(*configPath); err != nil {
		panic(err)
	}

	socks5 := &server.Socks5Server{
		DialFunc: codec.DialTrojan,
	}
	c := config.GetConfig()
	panic(socks5.ListenAndServe("tcp", c.LocalAddr))
}
