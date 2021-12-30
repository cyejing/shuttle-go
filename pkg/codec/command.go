package codec

import (
	"io"
)

//Command struct
type commandEnum byte

const (
	ConnectCE commandEnum = iota
	DialCE
	RespCE
)

type status int

const (
	SuccessStatus status = iota
	FailStatus
)

type req struct {
	commandEnum
	reqId int32
	len   int32
}

type ConnectCommand struct {
	*req
	name string
}

func (c ConnectCommand) Decode(r io.Reader) error{
	return nil
}

type DialCommand struct {
	*req
	*address
	body []byte
}

func (c DialCommand) Encode() ([]byte, error){
	return nil, nil
}


type RespCommand struct {
	*req
	status
	body []byte
}

var iotaReqId int32 = 0

func newReqId() int32 {
	iotaReqId += 1
	return iotaReqId
}

func NewDialCommand(body []byte) *DialCommand {
	return &DialCommand{
		req: &req{
			reqId:   newReqId(),
			commandEnum: DialCE,
			len:     int32(len(body)),
		},
		body: body,
	}
}
