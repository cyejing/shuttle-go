package operate

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/utils"
	"net"
)

func init() {
	registerOp(DialType, func() Operate {
		return &DialOP{
			ReqBase: new(ReqBase),
			Address: new(codec.Address),
		}
	})
}

type DialOP struct {
	*ReqBase
	*codec.Address
}

func (d *DialOP) Encode(buf *bytes.Buffer) error {
	addressBuf := bytes.NewBuffer([]byte{})
	err := d.Address.WriteTo(addressBuf)
	if err != nil {
		return utils.BaseErr("encode address err", err)
	}
	d.body = addressBuf.Bytes()
	reqBaseByte, err := d.ReqBase.Encode()
	if err != nil {
		return err
	}
	buf.Write(reqBaseByte)
	return nil
}

func (d *DialOP) Decode(buf *bufio.Reader) error {
	err := d.ReqBase.Decode(buf)
	if err != nil {
		return utils.BaseErr("connect command decode err", err)
	}
	addressBuf := bytes.NewBuffer(d.body)
	err = d.Address.ReadFrom(addressBuf)
	if err != nil {
		return utils.BaseErr("decode address err", err)
	}
	return nil
}

func (d *DialOP) Execute(ctx context.Context) error {
	//conn, err := net.Dial(d.Address.Network(), d.Address.String())
	//if err != nil {
	//	return err
	//}

	panic("implement me")
}

type ConnCtl struct {
	name       string
	address    *codec.Address
	dispatcher *Dispatcher

	raw        net.Conn
	rbuf       *bufio.Reader
	writeChan  chan []byte
}

func (c ConnCtl) Dial() error {
	conn, err := net.Dial(c.address.Network(), c.address.String())
	if err != nil {
		return utils.BaseErrf("dial address {} err", err, c.address.String())
	}
	c.raw = conn
	c.rbuf = bufio.NewReader(conn)
	c.dispatcher.DialMap.Store(c.name, c)
	return nil
}

func (c ConnCtl) Write(b []byte) error {
	_, err := c.raw.Write(b)
	if err != nil {
		return utils.BaseErrf("write conn {} err", err, c.address.String())
	}
	return nil
}

func (c ConnCtl) Read() error {
	buf := make([]byte,1024)
	for true {
		i, err := c.raw.Read(buf)
		if err != nil {
			return utils.BaseErrf("connCtl {} read err",err,c.name)
		}
		hex.Dump(buf[0:i])
		//c.dispatcher.Send(ExchangeOP buf)
	}
	return nil
}

func (c ConnCtl) Invalid(msg string) error {
	//c.dispatcher.Send(InvalidOP msg)
	return nil
}
