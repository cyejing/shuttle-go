package operate

import (
	"bufio"
	"bytes"
	"context"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/utils"
	"net"
)

type OpenProxy struct {
	*ReqBase
	ShipName   string
	RemoteAddr string
	LocalAddr  string
}

func init() {
	registerOp(OpenProxyType, func() Operate {
		return &OpenProxy{
			ReqBase: new(ReqBase),
		}
	})
}

func scanProxyConfig(d *Dispatcher) error {
	for _, ship := range client.GlobalConfig.Ships {
		d.Send(NewOpenProxyOP(ship.Name, ship.RemoteAddr, ship.LocalAddr))
	}
	return nil
}

func NewOpenProxyOP(shipName, remoteAddr, localAddr string) *OpenProxy {
	return &OpenProxy{
		ReqBase:    NewReqBase(OpenProxyType),
		ShipName:   shipName,
		RemoteAddr: remoteAddr,
		LocalAddr:  localAddr,
	}
}

var spliceStr = byte('\n')

func (o *OpenProxy) Encode(buf *bytes.Buffer) error {
	body := bytes.NewBuffer(make([]byte, 0))
	body.WriteString(o.ShipName)
	body.WriteByte(spliceStr)
	body.WriteString(o.RemoteAddr)
	body.WriteByte(spliceStr)
	body.WriteString(o.LocalAddr)
	body.WriteByte(spliceStr)

	o.body = body.Bytes()
	bs, err := o.ReqBase.Encode()
	if err != nil {
		return utils.BaseErr("openProxy op encode err", err)
	}
	buf.Write(bs)
	return nil
}

func (o *OpenProxy) Decode(buf *bufio.Reader) error {
	err := o.ReqBase.Decode(buf)
	if err != nil {
		return utils.BaseErr("openProxy op decode err", err)
	}
	reader := bufio.NewReader(bytes.NewReader(o.body))

	shipName, err := reader.ReadSlice(spliceStr)
	if err != nil {
		return utils.BaseErr("open proxy op read slice err", err)
	}
	o.ShipName = string(shipName[:len(shipName)-1])

	remoteAddr, err := reader.ReadSlice(spliceStr)
	if err != nil {
		return utils.BaseErr("open proxy op read slice err", err)
	}
	o.RemoteAddr = string(remoteAddr[:len(remoteAddr)-1])

	localAddr, err := reader.ReadSlice(spliceStr)
	if err != nil {
		return utils.BaseErr("open proxy op read slice err", err)
	}
	o.LocalAddr = string(localAddr[:len(localAddr)-1])

	return nil
}

func (o *OpenProxy) Execute(ctx context.Context) error {
	dispatcher, err := extractDispatcher(ctx)
	if err != nil {
		return err
	}
	go func() {
		err := NewProxyCtl(dispatcher.Name, o.ShipName, o.RemoteAddr, o.LocalAddr).Run()
		if err != nil {
			dispatcher.Send(NewRespOP(FailStatus, o.reqId, utils.BaseErr("new proxy ctl err",err).Error()))
		}else{
			dispatcher.Send(NewRespOP(SuccessStatus, o.reqId, "ok"))
		}
	}()
	return nil
}

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
		RemoteAddr:   remoteAddr,
		LocalAddr:    localAddr,
	}
}

func (p *ProxyCtl) Run() error {
	errChan := make(chan error)
	go func() {
		errChan <- p.run()
	}()
	return <- errChan
}

func (p *ProxyCtl) run() error {
	dispatcher := GetSerDispatcher(p.WormholeName)
	if dispatcher == nil {
		return utils.NewErrf("wormholeName %s does not exist", p.WormholeName)
	}
	ln, err := net.Listen("tcp", p.RemoteAddr)
	log.Infof("proxy listen at %s", p.RemoteAddr)
	if err != nil {
		return utils.BaseErr("proxy server ctl listen err",err)
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
	dispatcher := GetSerDispatcher(p.WormholeName)
	if dispatcher == nil {
		return utils.NewErrf("wormholeName %s does not exist", p.WormholeName)
	}
	exchangeCtl := NewExchangeCtl(p.ShipName, dispatcher, conn)
	addr, err := codec.NewAddressFromAddr("tcp", p.LocalAddr)
	if err != nil {
		return err
	}
	dispatcher.Send(NewDialOP(p.ShipName, addr))
	return exchangeCtl.Read()
}
