package operate

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRespOP_codec(t *testing.T) {
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

	rop, _ := typeMap[RespType]().(*RespOP)

	err = rop.Decode(bufio.NewReader(encodeByte))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, SuccessStatus, rop.Status)
	assert.Equal(t, uint32(8), rop.ReqId)
	assert.Equal(t, uint32(16), rop.Len)
	assert.Equal(t, len(body), len(rop.Body))
	assert.Equal(t, body[:], rop.Body)

}
