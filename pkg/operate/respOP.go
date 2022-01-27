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

//Status resp
type Status int

const (
	SuccessStatus Status = iota
	FailStatus
)

// RespOP struct
type RespOP struct {
	Type
	Status Status
	ReqId  uint32
	Len    uint32
	Body   []byte
}

func (rc *RespOP) RespCall() func(resp *RespOP) {
	return defaultCall
}

func init() {
	registerOp(RespType, func() Operate {
		return new(RespOP)
	})
}

func (rc *RespOP) Encode(buf *bytes.Buffer) error {
	buf.WriteByte(byte(rc.Type))
	buf.WriteByte(byte(rc.Status))
	reqIdByte := [4]byte{}
	binary.BigEndian.PutUint32(reqIdByte[:], rc.ReqId)
	buf.Write(reqIdByte[:])
	lenByte := [4]byte{}
	binary.BigEndian.PutUint32(lenByte[:], rc.Len)
	buf.Write(lenByte[:])

	buf.Write(rc.Body)
	return nil
}

func (rc *RespOP) Decode(buf *bufio.Reader) error {
	tb, err := codec.ReadByte(buf)
	if err != nil {
		return utils.BaseErr("req base decode fail", err)
	}
	rc.Type = Type(tb)

	statusByte, err := codec.ReadByte(buf)
	if err != nil {
		return utils.BaseErr("req base decode fail", err)
	}
	rc.Status = Status(statusByte)

	reqId, err := codec.ReadUint32(buf)
	if err != nil {
		return err
	}
	rc.ReqId = reqId

	bodyLen, err := codec.ReadUint32(buf)
	if err != nil {
		return err
	}
	rc.Len = bodyLen

	if bodyLen > 0 {
		body := make([]byte, bodyLen)
		_, err = io.ReadFull(buf, body)
		if err != nil {
			return utils.BaseErr("response command read Body fail", err)
		}
		rc.Body = body[:]
	}

	return nil
}

func (rc *RespOP) Execute(ctx context.Context) error {
	d, err := extractDispatcher(ctx)
	if err != nil {
		return err
	}
	if r, ok := d.ReqMap.LoadAndDelete(rc.ReqId); ok {
		if op, ok := r.(Operate); ok {
			op.RespCall()(rc)
		}
	}
	return nil
}

func (rc *RespOP) IsResponse() bool {
	return true
}

func (rc *RespOP) GetReqId() uint32 {
	return rc.ReqId
}

func NewRespOP(s Status, reqId uint32, msg string) *RespOP {
	buf := bytes.NewBufferString(msg)
	return &RespOP{
		Type: RespType,
		Status: s,
		ReqId:  reqId,
		Len:    uint32(buf.Len()),
		Body:   buf.Bytes(),
	}
}
