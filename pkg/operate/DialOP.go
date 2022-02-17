package operate

import (
	"bufio"
	"bytes"
	"context"
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
	nameLen uint32
	name    string
	*codec.Address
}

func NewDialOP(name string, address *codec.Address) *DialOP {
	return &DialOP{
		ReqBase: NewReqBase(DialType),
		name:    name,
		Address: address,
	}
}

func (d *DialOP) Encode(buf *bytes.Buffer) error {
	body := bytes.NewBuffer(make([]byte, 0))
	nameByte := []byte(d.name)
	body.Write(codec.EncodeUint32(uint32(len(nameByte))))
	body.Write(nameByte)
	err := d.Address.WriteTo(body)
	if err != nil {
		return utils.BaseErr("encode address err", err)
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
		return utils.BaseErr("connect command decode err", err)
	}
	d.nameLen = codec.DecodeUint32(d.body[:4])
	d.name = string(d.body[4 : 4+d.nameLen])
	addressBuf := bytes.NewBuffer(d.body[4+d.nameLen:])
	err = d.Address.ReadFrom(addressBuf)
	if err != nil {
		return utils.BaseErr("decode address err", err)
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
	ec := newExchangeCtl(d.name, dispatcher, conn)
	go func() {
		err := ec.Read()
		if err != nil {
			log.Warn(err)
		}
	}()

	//TODO
	return nil
}
