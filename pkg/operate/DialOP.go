package operate

import (
	"bufio"
	"bytes"
	"context"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/utils"
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
