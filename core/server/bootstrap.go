package server

import (
	config "github.com/cyejing/shuttle/pkg/config/server"
	"github.com/cyejing/shuttle/pkg/logger"
)


func Run(c *config.Config) {
	srv := &TLSServer{
		Cert:    c.Cert,
		Key:     c.Key,
		Handler: NewRouteMux(c),
	}
	ec := make(chan error, 2)
	go func() {
		err := srv.ListenAndServe(c.Addr)
		ec <- err
	}()
	go func() {
		err := srv.ListenAndServeTLS(c.SslAddr)
		ec <- err
	}()

	for i := 0; i < 2; i++ {
		e := <-ec
		logger.NewLog().Error(e)
	}
}

