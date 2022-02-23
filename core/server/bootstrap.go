package server

import (
	config "github.com/cyejing/shuttle/core/config/server"
	"github.com/cyejing/shuttle/core/filter"
	"github.com/cyejing/shuttle/pkg/logger"
)

var log = logger.NewLog()

func Run(c *config.Config) {
	srv := &TLSServer{
		Cert:    c.Cert,
		Key:     c.Key,
		Handler: filter.NewRouteMux(c),
	}
	ec := make(chan error, 2)
	if c.Addr != "" {
		go func() {
			err := srv.ListenAndServe(c.Addr)
			ec <- err
		}()
	}
	if c.SslAddr != "" && srv.Cert != "" && srv.Key != "" {
		go func() {
			err := srv.ListenAndServeTLS(c.SslAddr)
			ec <- err
		}()
	}

	e := <-ec
	log.Errorln("server exit", e)
}
