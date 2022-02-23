package operate

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnectOP_Codec(t *testing.T) {

	connectOP := NewConnectOP("test")
	encodeByte := bytes.NewBuffer([]byte{})
	err := connectOP.Encode(encodeByte)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(hex.Dump(encodeByte.Bytes()))

	cop := typeMap[ConnectType]().(*ConnectOp)
	err = cop.Decode(bufio.NewReader(encodeByte))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "test", cop.name)
}
