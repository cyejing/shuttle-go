package main

import (
	"flag"
	"github.com/cyejing/shuttle/core/server"
	config "github.com/cyejing/shuttle/pkg/config/server"
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

	srv := &server.TLSServer{
		Addr:    c.Addr,
		Cert:    c.Ssl.Cert,
		Key:     c.Ssl.Key,
		Handler: server.NewRouteMux(),
	}

	err = srv.ListenAndServeTLS()
	if err != nil {
		panic(err)
	}
}
