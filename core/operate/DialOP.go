package operate

import (
	"bufio"
	"bytes"
	"context"
	codec2 "github.com/cyejing/shuttle/core/codec"
	"github.com/cyejing/shuttle/pkg/errors"
	"net"
)

func init() {
	registerOp(DialType, func() Operate {
		return &DialOP{
			ReqBase: new(ReqBase),
			Address: new(codec2.Address),
		}
	})
}

type DialOP struct {
	*ReqBase
	nameLen uint32
	name    string
	*codec2.Address
}

func NewDialOP(name string, address *codec2.Address) *DialOP {
	return &DialOP{
		ReqBase: NewReqBase(DialType),
		name:    name,
		Address: address,
	}
}

func (d *DialOP) Encode(buf *bytes.Buffer) error {
	body := bytes.NewBuffer(make([]byte, 0))
	nameByte := []byte(d.name)
	body.Write(codec2.EncodeUint32(uint32(len(nameByte))))
	body.Write(nameByte)
	err := d.Address.WriteTo(body)
	if err != nil {
		return errors.BaseErr("encode address err", err)
	}
	d.body = body.Bytes()

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
		return errors.BaseErr("connect command decode err", err)
	}
	d.nameLen = codec2.DecodeUint32(d.body[:4])
	d.name = string(d.body[4 : 4+d.nameLen])
	addressBuf := bytes.NewBuffer(d.body[4+d.nameLen:])
	err = d.Address.ReadFrom(addressBuf)
	if err != nil {
		return errors.BaseErr("decode address err", err)
	}
	return nil
}

func (d *DialOP) Execute(ctx context.Context) error {
	dispatcher, err := extractDispatcher(ctx)
	if err != nil {
		return err
	}

	conn, err := net.Dial(d.Address.Network(), d.Address.String())
	if err != nil {
		return err
	}
	ec := NewExchangeCtl(d.name, dispatcher, conn)
	go func() {
		err := ec.Read()
		if err != nil {
			log.Warn(err)
		}
	}()

	return nil
}
