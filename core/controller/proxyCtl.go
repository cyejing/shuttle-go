package controller

import (
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/logger"
	"github.com/cyejing/shuttle/pkg/operate"
	"github.com/cyejing/shuttle/pkg/utils"
	"net"
)
var log = logger.NewLog()

type ProxyCtl struct {
	WormholeName string
	ShipName     string
	RemoteAddr   string
	LocalAddr    string
}

func NewProxyCtl(wormholeName, shipName, remoteAddr, localAddr string) *ProxyCtl {
	return &ProxyCtl{
		WormholeName: wormholeName,
		ShipName:     shipName,
		RemoteAddr:    remoteAddr,
		LocalAddr:    localAddr,
	}
}

func (p *ProxyCtl) Run() error {
	dispatcher := operate.GetSerDispatcher(p.WormholeName)
	if dispatcher == nil {
		return utils.NewErrf("wormholeName %s does not exist",p.WormholeName)
	}
	ln, err := net.Listen("tcp", p.RemoteAddr)
	log.Infof("proxy listen at %s", p.RemoteAddr)
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
			err := p.serveConn(conn)
			if err != nil {
				log.Errorln(utils.BaseErr("handle proxy fail", err))
				return
			}
		}()

	}
}

func (p *ProxyCtl) serveConn(conn net.Conn) error {
	dispatcher := operate.GetSerDispatcher(p.WormholeName)
	if dispatcher == nil {
		return utils.NewErrf("wormholeName %s does not exist",p.WormholeName)
	}
	exchangeCtl := operate.NewExchangeCtl(p.ShipName, dispatcher, conn)
	addr, err := codec.NewAddressFromAddr("tcp", p.LocalAddr)
	if err != nil {
		return err
	}
	dispatcher.Send(operate.NewDialOP(p.ShipName, addr))
	return exchangeCtl.Read()
}
