package operate

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReqBase_Codec(t *testing.T) {
	body := [16]byte{0xa, 0xa, 0xa}
	rb := &ReqBase{
		Type:  ConnectType,
		reqId: 8,
		len:   16,
		body:  body[:],
	}
	encodeByte, err := rb.Encode()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(hex.Dump(encodeByte))

	drb := ReqBase{}
	err = drb.Decode(bytes.NewBuffer(encodeByte))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, ConnectType, drb.Type)
	assert.Equal(t, uint32(8), drb.reqId)
	assert.Equal(t, uint32(16), drb.len)
	assert.Equal(t, len(body), len(drb.body))
	assert.Equal(t, body[:], drb.body)
}

func TestReqOp(t *testing.T) {
	t.SkipNow()

	nc := NewConnectOP("123")
	var ifc interface{} = nc
	if _, ok := ifc.(Operate); ok {
		fmt.Println("is Operate")
	}

	if _, ok := ifc.(ReqOperate); ok {
		fmt.Println("is ReqOperate")
	}

	if _, ok := ifc.(RespOperate); ok {
		fmt.Println("is RespOperate")
	}

	sbuf := bytes.NewBufferString("12345678")
	buf := make([]byte, 3)

	fmt.Println(sbuf.Bytes())
	for i := 0; i < 3; i++ {
		i, err := sbuf.Read(buf)
		if err != nil {
			return
		}
		fmt.Println(buf[0:i])
	}

}
