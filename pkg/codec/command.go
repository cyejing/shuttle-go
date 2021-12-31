package codec

import (
	"bytes"
	"encoding/binary"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
)

//Command struct
type commandEnum byte

const (
	ConnectCE commandEnum = iota
	DialCE
	RespCE
)

type ReqBase struct {
	commandEnum
	reqId    uint32
	len      uint32
	body     []byte
	respChan chan *RespCommand
}

func (rb *ReqBase) Decode(r io.Reader) error {
	ce, err := ReadByte(r)
	if err != nil {
		return utils.BaseErr("req base decode fail", err)
	}
	rb.commandEnum = commandEnum(ce)
	reqId, err := ReadUint32(r)
	if err != nil {
		return utils.BaseErr("req base decode fail", err)
	}
	rb.reqId = reqId
	bodyLen, err := ReadUint32(r)
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
	buf.WriteByte(byte(rb.commandEnum))
	reqIdByte := [4]byte{}
	binary.BigEndian.PutUint32(reqIdByte[:], rb.reqId)
	buf.Write(reqIdByte[:])
	lenByte := [4]byte{}
	binary.BigEndian.PutUint32(lenByte[:], rb.len)
	buf.Write(lenByte[:])

	buf.Write(rb.body)
	return buf.Bytes(), nil
}

type ConnectCommand struct {
	*ReqBase
	name string
}

func (c *ConnectCommand) Decode(r io.Reader) error {
	err := c.ReqBase.Decode(r)
	if err != nil {
		return utils.BaseErr("connect command decode fail", err)
	}
	c.name = string(c.body)
	return nil
}
func (c *ConnectCommand) Encode() ([]byte, error) {
	c.body = []byte(c.name)
	reqBaseByte, err := c.ReqBase.Encode()
	if err != nil {
		return nil, err
	}
	return reqBaseByte, nil
}

type DialCommand struct {
	*ReqBase
	*address
}

func (c DialCommand) Encode() ([]byte, error) {
	return nil, nil
}

//response status
type status int

const (
	SuccessStatus status = iota
	FailStatus
)

// RespCommand struct
type RespCommand struct {
	status
	reqId uint32
	len   uint32
	body  []byte
}

func (rc *RespCommand) Decode(r io.Reader) error {
	statusByte, err := ReadByte(r)
	if err != nil {
		return err
	}
	rc.status = status(statusByte)

	reqId, err := ReadUint32(r)
	if err != nil {
		return err
	}
	rc.reqId = reqId

	bodyLen, err := ReadUint32(r)
	if err != nil {
		return err
	}
	rc.len = bodyLen

	if bodyLen > 0 {
		body := make([]byte, bodyLen)
		_, err = io.ReadFull(r, body)
		if err != nil {
			return utils.BaseErr("response command read body fail", err)
		}
		rc.body = body[:]
	}

	return nil
}
func (rc *RespCommand) Encode() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteByte(byte(rc.status))
	reqIdByte := [4]byte{}
	binary.BigEndian.PutUint32(reqIdByte[:], rc.reqId)
	buf.Write(reqIdByte[:])
	lenByte := [4]byte{}
	binary.BigEndian.PutUint32(lenByte[:], rc.len)
	buf.Write(lenByte[:])

	buf.Write(rc.body)
	return buf.Bytes(), nil
}

var iotaReqId uint32 = 0

func newReqId() uint32 {
	iotaReqId += 1
	return iotaReqId
}

func NewDialCommand(body []byte) *DialCommand {
	return &DialCommand{
		ReqBase: &ReqBase{
			reqId:       newReqId(),
			commandEnum: DialCE,
			len:         uint32(len(body)),
			body:        body,
		},
	}
}

func NewConnectCommand(name string) *ConnectCommand {
	nameByte := []byte(name)
	return &ConnectCommand{
		ReqBase: &ReqBase{
			commandEnum: ConnectCE,
			reqId:       newReqId(),
			len:         uint32(len(nameByte)),
			body:        nameByte,
		},
		name:    name,
	}
}
