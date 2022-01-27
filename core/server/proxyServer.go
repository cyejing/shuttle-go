package server

import (
	"net"
)

type ProxyServer struct {
}

func (p *ProxyServer) ListenAndServe(network, addr string) error {
	ln, err := net.Listen(network, addr)
	log.Infof("proxy listen at %s", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Errorln("proxy accept conn fail", err)
		}

		go func() {
			defer conn.Close()
			err := p.ServeConn(conn)
			if err != nil {
				log.Errorln("handle proxy fail ", err)
				return
			}
		}()

	}
}

func (p *ProxyServer) ServeConn(c net.Conn) error {
	return nil
}
