package server

import (
	"crypto/tls"
	"github.com/cyejing/shuttle/core/channel"
	"github.com/cyejing/shuttle/core/config"
	"github.com/cyejing/shuttle/pkg/errors"
	"github.com/cyejing/shuttle/pkg/logger"
	"net"
)

var log = logger.NewLog()

func Run(c config.ServerConfig) {
	addrLen := len(c.Addrs)
	ec := make(chan error, addrLen)
	for _, addr := range c.Addrs {
		NewTlsServer(c, addr.Addr, addr.Cert, addr.Key).Run(ec)
	}
	for i := 0; i < addrLen; i++ {
		err := <-ec
		log.Errorln(err)
	}
}

type TlsServer struct {
	config config.ServerConfig
	Addr   string
	Cert   string
	Key    string
}

func NewTlsServer(c config.ServerConfig, addr, cert, key string) *TlsServer {
	return &TlsServer{
		config: c,
		Addr:   addr,
		Cert:   cert,
		Key:    key,
	}
}

func (t *TlsServer) Run(ec chan error) {
	ec <- t.ServAndListen()
}

func (t *TlsServer) ServAndListen() error {
	cert, err := tls.LoadX509KeyPair(t.Cert, t.Key)
	if err != nil {
		return errors.BaseErr("tls load cert or key file failed", err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", t.Addr, config)
	if err != nil {
		return errors.BaseErr("listen addr failed", err)
	}

	return t.Serv(ln)
}

func (t *TlsServer) Serv(ln net.Listener) error {
	for true {
		raw, err := ln.Accept()
		if err != nil {
			log.Error(err)
		}

		go channel.NewPeekChannel(raw, t.config).Run()
	}

	return nil
}
