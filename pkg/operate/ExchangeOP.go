package operate

import (
	"bufio"
	"bytes"
	"context"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/utils"
	"net"
)

type ExchangeOP struct {
	*ReqBase
	nameLen uint32
	name    string
	data    []byte
}

func init() {
	registerOp(ExchangeType, func() Operate {
		return &ExchangeOP{
			ReqBase: new(ReqBase),
		}
	})
}

func NewExchangeOP(name string, data []byte) *ExchangeOP {
	return &ExchangeOP{
		ReqBase: NewReqBase(ExchangeType),
		nameLen: 0,
		name:    name,
		data:    data,
	}
}

func (e *ExchangeOP) Encode(buf *bytes.Buffer) error {
	body := bytes.NewBuffer(make([]byte, 0))
	nameByte := []byte(e.name)
	body.Write(codec.EncodeUint32(uint32(len(nameByte))))
	body.Write(nameByte)
	body.Write(e.data)
	e.body = body.Bytes()
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
	e.nameLen = codec.DecodeUint32(e.body[:4])
	e.name = string(e.body[4 : 4+e.nameLen])
	e.data = e.body[4+e.nameLen:]
	return nil
}

func (e *ExchangeOP) Execute(ctx context.Context) error {
	d, err := extractDispatcher(ctx)
	if err != nil {
		return err
	}
	if exchangeCtl, ok := d.LoadExchange(e.name); ok {
		err := exchangeCtl.Write(e.data)
		if err != nil {
			return utils.BaseErr("exchange ctl write data err", err)
		}
	}
	return nil
}

type ExchangeCtl interface {
	Write(b []byte) error
}

type ExchangeCtlStu struct {
	name       string
	dispatcher *Dispatcher

	raw net.Conn
}

func NewExchangeCtl(name string, d *Dispatcher, raw net.Conn) *ExchangeCtlStu {
	ecs := &ExchangeCtlStu{
		name:       name,
		dispatcher: d,
		raw:        raw,
	}
	d.exchangeMap.Store(name, ecs)
	return ecs
}

func (c *ExchangeCtlStu) Write(b []byte) error {
	_, err := c.raw.Write(b)
	if err != nil {
		return utils.BaseErrf("write conn %v err", err, c.raw)
	}
	return nil
}

func (c *ExchangeCtlStu) Read() error {
	for true {
		// send OP, cannot reuse buf
		buf := make([]byte, 1024 * 4)
		i, err := c.raw.Read(buf)
		if err != nil {
			return utils.BaseErrf("connCtl %s read err", err, c.name)
		}
		op := NewExchangeOP(c.name, buf[0:i])
		c.dispatcher.Send(op)
	}
	return nil
}

func (c *ExchangeCtlStu) Invalid(msg string) {
	c.dispatcher.DeleteExchange(c.name)
	log.Warnf("exchange ctl stu invalid name:%s, msg:%s", c.name, msg)
	//c.dispatcher.Send(InvalidOP msg) TODO
}
