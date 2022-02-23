package operate

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/cyejing/shuttle/core/codec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDialOPCodec(t *testing.T) {

	addr, err := codec.NewAddressFromAddr("tcp", "127.0.0.1:4080")
	if err != nil {
		t.Error(err)
	}
	dialOP := NewDialOP("test", addr)
	encodeByte := bytes.NewBuffer([]byte{})

	err = dialOP.Encode(encodeByte)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(hex.Dump(encodeByte.Bytes()))

	dop := typeMap[DialType]().(*DialOP)
	err = dop.Decode(bufio.NewReader(encodeByte))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "test", dop.name)
	assert.Equal(t, codec.IPv4,dop.Address.AddressType)
	assert.Equal(t, "127.0.0.1:4080",dop.Address.String())
}
