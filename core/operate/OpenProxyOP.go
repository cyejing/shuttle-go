package operate

import (
	"bufio"
	"bytes"
	"context"
	"github.com/cyejing/shuttle/core/codec"
	"github.com/cyejing/shuttle/core/config/client"
	"github.com/cyejing/shuttle/pkg/errors"
	"github.com/cyejing/shuttle/pkg/utils"
	"net"
	"strings"
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
		return errors.BaseErr("openProxy op encode err", err)
	}
	buf.Write(bs)
	return nil
}

func (o *OpenProxy) Decode(buf *bufio.Reader) error {
	err := o.ReqBase.Decode(buf)
	if err != nil {
		return errors.BaseErr("openProxy op decode err", err)
	}
	reader := bufio.NewReader(bytes.NewReader(o.body))

	shipName, err := reader.ReadSlice(spliceStr)
	if err != nil {
		return errors.BaseErr("open proxy op read slice err", err)
	}
	o.ShipName = string(shipName[:len(shipName)-1])

	remoteAddr, err := reader.ReadSlice(spliceStr)
	if err != nil {
		return errors.BaseErr("open proxy op read slice err", err)
	}
	o.RemoteAddr = string(remoteAddr[:len(remoteAddr)-1])

	localAddr, err := reader.ReadSlice(spliceStr)
	if err != nil {
		return errors.BaseErr("open proxy op read slice err", err)
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
		err := NewProxyCtl(dispatcher, o.ShipName, o.RemoteAddr, o.LocalAddr).Run()
		if err != nil {
			dispatcher.Send(NewRespOP(FailStatus, o.reqId, errors.BaseErr("new proxy ctl err", err).Error()))
		} else {
			dispatcher.Send(NewRespOP(SuccessStatus, o.reqId, "ok"))
		}
	}()
	return nil
}

type ProxyCtl struct {
	dispatcher *Dispatcher
	ShipName   string
	RemoteAddr string
	LocalAddr  string
	ctx        context.Context
	stopFunc   context.CancelFunc
}

func NewProxyCtl(dispatcher *Dispatcher, shipName, remoteAddr, localAddr string) *ProxyCtl {
	ctx, stopFunc := context.WithCancel(context.Background())
	proxyCtl := &ProxyCtl{
		dispatcher: dispatcher,
		ShipName:   shipName,
		RemoteAddr: remoteAddr,
		LocalAddr:  localAddr,
		ctx:        ctx,
		stopFunc:   stopFunc,
	}
	dispatcher.ProxyMap.Store(shipName, proxyCtl)
	return proxyCtl
}

func (p *ProxyCtl) Run() error {
	errChan := make(chan error)
	go func() {
		errChan <- p.run()
	}()
	return <-errChan
}

func (p *ProxyCtl) Stop() {
	p.stopFunc()
}

func (p *ProxyCtl) run() error {
	log.Infof("open proxy at %s", p.RemoteAddr)
	ln, err := net.Listen("tcp", p.RemoteAddr)
	log.Infof("proxy [%s] listen at %s", p.ShipName, p.RemoteAddr)
	if err != nil {
		return errors.BaseErr("proxy server ctl listen err", err)
	}
	for {
		connChan := make(chan net.Conn)
		go func() {
			conn, err := ln.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					log.Errorln("proxy accept conn fail", err)
				}
			}
			connChan <- conn
		}()
		select {
		case <-p.ctx.Done():
			log.Infof("proxy [%s] server stop, remote[%s]", p.ShipName, p.RemoteAddr)
			ln.Close()
			return nil
		case conn := <-connChan:
			go func() {
				defer conn.Close()
				err := p.serveConn(conn)
				if err != nil {
					log.Warn(errors.BaseErr("handle proxy fail", err))
					return
				}
			}()
		}
	}
}

func (p *ProxyCtl) serveConn(conn net.Conn) error {
	uniqueId := utils.GenUniqueId()
	exchangeCtl := NewExchangeCtl(uniqueId, p.dispatcher, conn)
	addr, err := codec.NewAddressFromAddr("tcp", p.LocalAddr)
	if err != nil {
		return err
	}
	p.dispatcher.Send(NewDialOP(uniqueId, addr))
	return exchangeCtl.Read()
}
