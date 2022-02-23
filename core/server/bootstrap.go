package server

import (
	config "github.com/cyejing/shuttle/core/config/server"
	"github.com/cyejing/shuttle/core/filter"
	"github.com/cyejing/shuttle/pkg/logger"
)

var log = logger.NewLog()

func Run(c *config.Config) {
	addrLen := len(c.Addrs)
	ec := make(chan error, addrLen)
	for _, addr := range c.Addrs {
		NewHttpServer(addr.Addr, addr.Cert, addr.Key, filter.NewRouteMux(c)).Run(ec)
	}
	for i := 0; i < addrLen; i++ {
		err := <-ec
		log.Errorln(err)
	}
}
