package operate

import (
	"bufio"
	"bytes"
	"context"
	"github.com/cyejing/shuttle/core/codec"
	"github.com/cyejing/shuttle/pkg/errors"
	"net"
)

type ExchangeOP struct {
	*ReqBase
	nameLen uint32
	name    string
	invalid bool
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
		invalid: false,
		data:    data,
	}
}

func (e *ExchangeOP) Encode(buf *bytes.Buffer) error {
	body := bytes.NewBuffer(make([]byte, 0))
	nameByte := []byte(e.name)
	body.Write(codec.EncodeUint32(uint32(len(nameByte))))
	body.Write(nameByte)
	if e.invalid {
		body.WriteByte(byte(1))
	} else {
		body.WriteByte(byte(0))
	}
	body.Write(e.data)
	e.body = body.Bytes()
	bs, err := e.ReqBase.Encode()
	if err != nil {
		return errors.BaseErr("exchange op encode err", err)
	}
	buf.Write(bs)
	return nil
}

func (e *ExchangeOP) Decode(buf *bufio.Reader) error {
	err := e.ReqBase.Decode(buf)
	if err != nil {
		return errors.BaseErr("exchange op decode err", err)
	}
	e.nameLen = codec.DecodeUint32(e.body[:4])
	e.name = string(e.body[4 : 4+e.nameLen])
	i := e.body[4+e.nameLen]
	if i == 0 {
		e.invalid = false
	} else {
		e.invalid = true
	}
	e.data = e.body[5+e.nameLen:]
	return nil
}

func (e *ExchangeOP) Execute(ctx context.Context) error {
	d, err := extractDispatcher(ctx)
	if err != nil {
		return err
	}
	if exchangeCtl, ok := d.LoadExchange(e.name); ok {
		if e.invalid {
			exchangeCtl.Close()
			return nil
		}
		err := exchangeCtl.Write(e.data)
		if err != nil {
			exchangeCtl.SendInvalid()
		}
	}
	return nil
}

type ExchangeCtl interface {
	Write(b []byte) error
	Close()
	SendInvalid()
}

type ExchangeCtlStu struct {
	Name       string
	dispatcher *Dispatcher

	Raw net.Conn
}

func NewExchangeCtl(name string, d *Dispatcher, raw net.Conn) *ExchangeCtlStu {
	ecs := &ExchangeCtlStu{
		Name:       name,
		dispatcher: d,
		Raw:        raw,
	}
	d.ExchangeMap.Store(name, ecs)
	return ecs
}

func (c *ExchangeCtlStu) Write(b []byte) error {
	_, err := c.Raw.Write(b)
	if err != nil {
		return errors.BaseErrf("write conn %v err", err, c.Raw)
	}
	return nil
}

func (c *ExchangeCtlStu) Read() error {
	for true {
		// send OP, cannot reuse buf
		buf := make([]byte, 1024*4)
		i, err := c.Raw.Read(buf)
		if err != nil {
			c.SendInvalid()
			if errors.IsNetErr(err) {
				return nil
			}
			return errors.BaseErrf("connCtl %s read err", err, c.Name)
		}
		op := NewExchangeOP(c.Name, buf[0:i])
		c.dispatcher.Send(op)
	}
	return nil
}

func (c *ExchangeCtlStu) Close() {
	c.dispatcher.DeleteExchange(c.Name)
	log.Infof("exchange ctl stu close name:%s", c.Name)
	c.Raw.Close()
}

func (c *ExchangeCtlStu) SendInvalid() {
	if _, ok := c.dispatcher.LoadExchange(c.Name); ok {
		invalidOP := NewExchangeOP(c.Name, nil)
		invalidOP.invalid = true
		c.dispatcher.Send(invalidOP)
		c.Close()
	}
}
