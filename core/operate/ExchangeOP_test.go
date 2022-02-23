package operate

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExchangeOPCodec(t *testing.T) {
	data := []byte{0xa, 0xa, 0xa}
	exchangeOP := NewExchangeOP("test", data)

	encodeByte := bytes.NewBuffer([]byte{})

	err := exchangeOP.Encode(encodeByte)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(hex.Dump(encodeByte.Bytes()))

	eop:= typeMap[ExchangeType]().(*ExchangeOP)
	err = eop.Decode(bufio.NewReader(encodeByte))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "test", eop.name)
	assert.Equal(t, uint32(len([]byte("test"))), eop.nameLen)
	assert.Equal(t, data, eop.data)
}
