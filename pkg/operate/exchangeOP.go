package operate

import (
	"bufio"
	"bytes"
	"context"
	"github.com/cyejing/shuttle/pkg/utils"
)

type ExchangeOp struct {
	*ReqBase
	name string
}

func init() {
	registerOp(ExchangeType, func() Operate {
		return &ExchangeOp{
			ReqBase: new(ReqBase),
		}
	})
}

func (c *ExchangeOp) Encode(buf *bytes.Buffer) error {
	c.body = []byte(c.name)
	reqBaseByte, err := c.ReqBase.Encode()
	if err != nil {
		return err
	}
	buf.Write(reqBaseByte)
	return nil
}

func (c *ExchangeOp) Decode(buf *bufio.Reader) error {
	err := c.ReqBase.Decode(buf)
	if err != nil {
		return utils.BaseErr("connect command decode fail", err)
	}
	c.name = string(c.body)
	return nil
}

func (c *ExchangeOp) Execute(ctx context.Context) error {
	d, err := extractDispatcher(ctx)
	if err != nil {
		return err
	}
	d.Wormhole.Name = c.name
	log.Infof("wormhole exchange name:%s", c.name)

	d.Send(NewRespOP(SuccessStatus, c.reqId, "ok"))
	return nil
}

func (c *ExchangeOp) IsResponse() bool {
	return false
}

func (c *ExchangeOp) GetReqId() uint32 {
	return c.reqId
}

func (c *ExchangeOp) RespCall() func(resp *RespOP) {
	return defaultCall
}

func NewExchangeOP(name string, call func(resp *RespOP)) *ExchangeOp {
	nameByte := []byte(name)
	if call == nil {
		call = defaultCall
	}
	return &ExchangeOp{
		ReqBase: &ReqBase{
			Type:     ExchangeType,
			reqId:    newReqId(),
			len:      uint32(len(nameByte)),
			body:     nameByte,
		},
		name: name,
	}
}
