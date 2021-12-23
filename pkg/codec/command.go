package codec

//Command struct
type command byte

const (
	DialCommand command = iota
	RespCommand
)

type req struct {
	command
	reqId int32
	len   int32
}

type DialCommandS struct {
	*req
	*address
	body []byte
}

type RespCommandS struct {
	*req
	body []byte
}

var iotaReqId int32 = 0

func newReqId() int32 {
	iotaReqId += 1
	return iotaReqId
}

func NewDialCommand(body []byte) *DialCommandS {
	return &DialCommandS{
		req: &req{
			reqId:   newReqId(),
			command: DialCommand,
			len:     int32(len(body)),
		},
		body: body,
	}
}
