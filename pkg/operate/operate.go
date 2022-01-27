package operate

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
)

//Type struct
type Type byte

const (
	ExchangeType Type = iota
	DialType
	RespType
)

type Operate interface {

	IsResponse() bool

	GetReqId() uint32

	RespCall() func(resp *RespOP)

	Encode(buf *bytes.Buffer) error

	Decode(buf *bufio.Reader) error

	Execute(ctx context.Context) error
}

type ReqBase struct {
	Type
	reqId    uint32
	len      uint32
	body     []byte
}

func (rb *ReqBase) Decode(r io.Reader) error {
	tb, err := codec.ReadByte(r)
	if err != nil {
		return utils.BaseErr("req base decode fail", err)
	}
	rb.Type = Type(tb)
	reqId, err := codec.ReadUint32(r)
	if err != nil {
		return utils.BaseErr("req base decode fail", err)
	}
	rb.reqId = reqId
	bodyLen, err := codec.ReadUint32(r)
	if err != nil {
		return utils.BaseErr("req base decode fail", err)
	}
	rb.len = bodyLen
	if bodyLen > 0 {
		bodyBytes := make([]byte, bodyLen)
		_, err = io.ReadFull(r, bodyBytes)
		if err != nil {
			return utils.BaseErr("req base decode fail", err)
		}
		rb.body = bodyBytes
	}
	return nil
}

func (rb *ReqBase) Encode() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteByte(byte(rb.Type))
	reqIdByte := [4]byte{}
	binary.BigEndian.PutUint32(reqIdByte[:], rb.reqId)
	buf.Write(reqIdByte[:])
	lenByte := [4]byte{}
	binary.BigEndian.PutUint32(lenByte[:], rb.len)
	buf.Write(lenByte[:])

	buf.Write(rb.body)
	return buf.Bytes(), nil
}

var iotaReqId uint32 = 0

func newReqId() uint32 {
	iotaReqId += 1
	return iotaReqId
}

func defaultCall(resp *RespOP) {
	log.Debugf("reqId %v have response Status:%v msg:%s", resp.ReqId, resp.Status, string(resp.Body))
}
