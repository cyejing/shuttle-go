package operate

import (
	"bufio"
	"bytes"
	"context"
	"github.com/cyejing/shuttle/pkg/utils"
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
		return utils.BaseErr("connect command decode fail", err)
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

func NewConnectOP(name string) *ConnectOp {
	nameByte := []byte(name)
	return &ConnectOp{
		ReqBase: &ReqBase{
			Type:  ConnectType,
			reqId: newReqId(),
			len:   uint32(len(nameByte)),
			body:  nameByte,
		},
		name: name,
	}
}
