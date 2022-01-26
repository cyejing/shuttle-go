package codec

import (
	"bytes"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReqBase_Decode(t *testing.T) {
	body := [16]byte{0xa, 0xa, 0xa}
	rb := &ReqBase{
		commandEnum: ExchangeCE,
		reqId:       8,
		len:         16,
		body:        body[:],
	}
	encodeByte, err := rb.Encode()
	if err != nil {
		t.Error(err)
	}
	t.Log(hex.Dump(encodeByte))

	drb := ReqBase{}
	err = drb.Decode(bytes.NewBuffer(encodeByte))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, ExchangeCE, drb.commandEnum)
	assert.Equal(t, uint32(8), drb.reqId)
	assert.Equal(t, uint32(16), drb.len)
	assert.Equal(t, len(body), len(drb.body))
	assert.Equal(t, body[:], drb.body)
}

func TestRespCommand_Decode(t *testing.T) {
	body := [16]byte{0xa, 0xa, 0xa}
	rc := &RespCommand{
		status: SuccessStatus,
		ReqId:  8,
		Len:    16,
		Body:   body[:],
	}
	encodeByte, err := rc.Encode()
	if err != nil {
		t.Error(err)
	}
	t.Log(hex.Dump(encodeByte))

	drc := &RespCommand{}
	err = drc.Decode(bytes.NewReader(encodeByte))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, SuccessStatus, drc.status)
	assert.Equal(t, uint32(8), drc.ReqId)
	assert.Equal(t, uint32(16), drc.Len)
	assert.Equal(t, len(body), len(drc.Body))
	assert.Equal(t, body[:], drc.Body)
}
