package core

import (
	"github.com/cyejing/shuttle/pkg"
	"github.com/cyejing/shuttle/pkg/handler"
	"net"
)

func NewEcho(port int) {
	server, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
	if err != nil {
		panic(err)
	}
	for {
		conn, err := server.AcceptTCP()
		if err != nil {
			return
		}
		go pkg.NewTCP(conn, []pkg.Handler{&handler.Echo{}}).Handler()
	}
}
