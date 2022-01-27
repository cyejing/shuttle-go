package operate

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRespCommand_Decode(t *testing.T) {
	body := [16]byte{0xa, 0xa, 0xa}
	rc := &RespOP{
		Status: SuccessStatus,
		ReqId:  8,
		Len:    16,
		Body:   body[:],
	}
	encodeByte := bytes.NewBuffer([]byte{})
	err := rc.Encode(encodeByte)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(hex.Dump(encodeByte.Bytes()))


	drc := &RespOP{}
	err = drc.Decode(bufio.NewReader(encodeByte))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, SuccessStatus, drc.Status)
	assert.Equal(t, uint32(8), drc.ReqId)
	assert.Equal(t, uint32(16), drc.Len)
	assert.Equal(t, len(body), len(drc.Body))
	assert.Equal(t, body[:], drc.Body)
}

