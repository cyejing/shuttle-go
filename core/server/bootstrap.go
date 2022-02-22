package server

import (
	"github.com/cyejing/shuttle/core/filter"
	config "github.com/cyejing/shuttle/pkg/config/server"
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
	go func() {
		err := srv.ListenAndServe(c.Addr)
		ec <- err
	}()
	if srv.Cert!="" && srv.Key!= "" {
		go func() {
			err := srv.ListenAndServeTLS(c.SslAddr)
			ec <- err
		}()
	}

	e := <-ec
	logger.NewLog().Error(e)
	log.Infof("server exit")
}

