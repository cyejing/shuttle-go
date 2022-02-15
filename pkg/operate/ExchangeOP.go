package operate

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"github.com/cyejing/shuttle/pkg/utils"
	"net"
)

type ExchangeOP struct {
	*ReqBase
	nameLen uint32
	name    string
	data    []byte
}

func (e *ExchangeOP) Encode(buf *bytes.Buffer) error {
	bs, err := e.ReqBase.Encode()
	if err != nil {
		return utils.BaseErr("exchange op encode err", err)
	}
	buf.Write(bs)
	return nil
}

func (e *ExchangeOP) Decode(buf *bufio.Reader) error {
	err := e.ReqBase.Decode(buf)
	if err != nil {
		return utils.BaseErr("exchange op decode err", err)
	}
	return nil
}

func (e *ExchangeOP) Execute(ctx context.Context) error {
	d, err := extractDispatcher(ctx)
	if err != nil {
		return err
	}
	if exchangeCtl, ok := d.LoadExchange(e.name); ok {
		println(exchangeCtl)
	}
	return nil
}

type ExchangeConn interface {
}

type ExchangeCtl struct {
	name       string
	dispatcher *Dispatcher

	raw net.Conn
}

func (c *ExchangeCtl) Write(b []byte) error {
	_, err := c.raw.Write(b)
	if err != nil {
		return utils.BaseErrf("write conn {} err", err, c.raw)
	}
	return nil
}

func (c *ExchangeCtl) Read() error {
	buf := make([]byte, 1024)
	for true {
		i, err := c.raw.Read(buf)
		if err != nil {
			return utils.BaseErrf("connCtl {} read err", err, c.name)
		}
		hex.Dump(buf[0:i])
		//c.dispatcher.Send(ExchangeOP buf)
	}
	return nil
}

func (c *ExchangeCtl) Invalid(msg string) error {
	c.dispatcher.DeleteExchange(c.name)
	//c.dispatcher.Send(InvalidOP msg)
	return nil
}
