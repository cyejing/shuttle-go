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

func (c DialCommandS) Encode() ([]byte, error){
	return nil, nil
}

type status int

const (
	Success status = iota
	Fail
)

type RespCommandS struct {
	*req
	status
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
