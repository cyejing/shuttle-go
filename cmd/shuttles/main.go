package main

import (
	"flag"
	"github.com/cyejing/shuttle/core/server"
	config "github.com/cyejing/shuttle/pkg/config/server"
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

	srv := &server.TLSServer{
		Cert:    c.Cert,
		Key:     c.Key,
		Handler: server.NewRouteMux(c),
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
