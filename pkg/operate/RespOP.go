package operate

import (
	"bufio"
	"bytes"
	"context"
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

func init() {
	registerOp(RespType, func() Operate {
		return new(RespOP)
	})
}

func (r *RespOP) GetRespStatus() Status {
	return r.Status
}

func (r *RespOP) Encode(buf *bytes.Buffer) error {
	buf.WriteByte(byte(r.Type))
	buf.WriteByte(byte(r.Status))
	reqIdByte := codec.EncodeUint32(r.ReqId)
	buf.Write(reqIdByte)
	lenByte := codec.EncodeUint32(r.Len)
	buf.Write(lenByte)

	buf.Write(r.Body)
	return nil
}

func (r *RespOP) Decode(buf *bufio.Reader) error {
	tb, err := codec.ReadByte(buf)
	if err != nil {
		return utils.BaseErr("req base decode fail", err)
	}
	r.Type = Type(tb)

	statusByte, err := codec.ReadByte(buf)
	if err != nil {
		return utils.BaseErr("req base decode fail", err)
	}
	r.Status = Status(statusByte)

	reqId, err := codec.ReadUint32(buf)
	if err != nil {
		return err
	}
	r.ReqId = reqId

	bodyLen, err := codec.ReadUint32(buf)
	if err != nil {
		return err
	}
	r.Len = bodyLen

	if bodyLen > 0 {
		body := make([]byte, bodyLen)
		_, err = io.ReadFull(buf, body)
		if err != nil {
			return utils.BaseErr("response command read Body fail", err)
		}
		r.Body = body[:]
	}

	return nil
}

func (r *RespOP) Execute(ctx context.Context) error {
	d, err := extractDispatcher(ctx)
	if err != nil {
		return err
	}
	if req, ok := d.LoadReq(r.ReqId); ok {
		req.RespCall()(req.GetReqBase(), r)
	}
	return nil
}

func NewRespOP(s Status, reqId uint32, msg string) *RespOP {
	buf := bytes.NewBufferString(msg)
	return &RespOP{
		Type:   RespType,
		Status: s,
		ReqId:  reqId,
		Len:    uint32(buf.Len()),
		Body:   buf.Bytes(),
	}
}
