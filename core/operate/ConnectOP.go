package operate

import (
	"bufio"
	"bytes"
	"context"
	"github.com/cyejing/shuttle/pkg/errors"
)

type ConnectOp struct {
	*ReqBase
	name string
}

func init() {
	registerOp(ConnectType, func() Operate {
		return &ConnectOp{
			ReqBase: new(ReqBase),
		}
	})
}

func NewConnectOP(name string) *ConnectOp {
	return &ConnectOp{
		ReqBase: NewReqBase(ConnectType),
		name:    name,
	}
}

func (c *ConnectOp) Encode(buf *bytes.Buffer) error {
	c.body = []byte(c.name)
	reqBaseByte, err := c.ReqBase.Encode()
	if err != nil {
		return err
	}
	buf.Write(reqBaseByte)
	return nil
}

func (c *ConnectOp) Decode(buf *bufio.Reader) error {
	err := c.ReqBase.Decode(buf)
	if err != nil {
		return errors.BaseErr("connect command decode fail", err)
	}
	c.name = string(c.body)
	return nil
}

func (c *ConnectOp) Execute(ctx context.Context) error {
	d, err := extractDispatcher(ctx)
	if err != nil {
		return err
	}
	log.Infof("wormhole connect name:%s", c.name)

	d.Send(NewRespOP(SuccessStatus, c.reqId, "ok"))
	return nil
}
