package operate

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReqBase_Decode(t *testing.T) {
	body := [16]byte{0xa, 0xa, 0xa}
	rb := &ReqBase{
		Type:  ExchangeType,
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
	assert.Equal(t, ExchangeType, drb.Type)
	assert.Equal(t, uint32(8), drb.reqId)
	assert.Equal(t, uint32(16), drb.len)
	assert.Equal(t, len(body), len(drb.body))
	assert.Equal(t, body[:], drb.body)
}
