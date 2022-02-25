package operate

import (
	"bufio"
	"bytes"
	"context"
	"github.com/cyejing/shuttle/core/codec"
	"github.com/cyejing/shuttle/pkg/errors"
	"io"
)

//Type struct
type Type byte

const (
	ConnectType Type = iota
	RespType
	DialType
	ExchangeType
	OpenProxyType
)

//Operate interface more
type Operate interface {
	Encode(buf *bytes.Buffer) error

	Decode(buf *bufio.Reader) error

	Execute(ctx context.Context) error
}

type ReqOperate interface {
	Operate

	GetReqBase() *ReqBase

	RespCall() func(req *ReqBase, resp *RespOP)

	WaitResp() *RespOP
}

type RespOperate interface {
	Operate

	GetRespStatus() Status
}

//ReqBase struct
type ReqBase struct {
	Type
	reqId    uint32
	len      uint32
	body     []byte
	respChan chan *RespOP
	respCall func(req *ReqBase, resp *RespOP)
}

func NewReqBase(t Type) *ReqBase {
	return &ReqBase{
		Type:     t,
		reqId:    newReqId(),
		respChan: make(chan *RespOP),
	}
}

func (rb *ReqBase) Decode(r io.Reader) error {
	tb, err := codec.ReadByte(r)
	if err != nil {
		return errors.BaseErr("req base decode fail", err)
	}
	rb.Type = Type(tb)
	reqId, err := codec.ReadUint32(r)
	if err != nil {
		return errors.BaseErr("req base decode fail", err)
	}
	rb.reqId = reqId
	bodyLen, err := codec.ReadUint32(r)
	if err != nil {
		return errors.BaseErr("req base decode fail", err)
	}
	rb.len = bodyLen
	if bodyLen > 0 {
		bodyBytes := make([]byte, bodyLen)
		_, err = io.ReadFull(r, bodyBytes)
		if err != nil {
			return errors.BaseErr("req base decode fail", err)
		}
		rb.body = bodyBytes
	}
	return nil
}

func (rb *ReqBase) Encode() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteByte(byte(rb.Type))
	reqIdByte := codec.EncodeUint32(rb.reqId)
	buf.Write(reqIdByte)
	lenByte := codec.EncodeUint32(uint32(len(rb.body)))
	buf.Write(lenByte)

	buf.Write(rb.body)
	return buf.Bytes(), nil
}

func (rb *ReqBase) GetReqBase() *ReqBase {
	return rb
}

func (rb *ReqBase) RespCall() func(req *ReqBase, resp *RespOP) {
	if rb.respCall == nil {
		return func(req *ReqBase, resp *RespOP) {}
	} else {
		return rb.respCall
	}
}

func (rb ReqBase) WaitResp() *RespOP {
	rb.respCall = func(req *ReqBase, resp *RespOP) {
		req.respChan <- resp
	}
	return <-rb.respChan
}

//more func
var iotaReqId uint32 = 0

func newReqId() uint32 {
	iotaReqId += 1
	return iotaReqId
}
