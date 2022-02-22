package operate

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenProxy_Codec(t *testing.T) {
	proxyOP := NewOpenProxyOP("test", "127.0.0.1:2121", "127.0.0.1:2122")
	encodeByte := bytes.NewBuffer([]byte{})

	err := proxyOP.Encode(encodeByte)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(hex.Dump(encodeByte.Bytes()))

	pop:=typeMap[OpenProxyType]().(*OpenProxy)

	err = pop.Decode(bufio.NewReader(encodeByte))

	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "test", pop.ShipName)
	assert.Equal(t, "127.0.0.1:2121", pop.RemoteAddr)
	assert.Equal(t, "127.0.0.1:2122", pop.LocalAddr)

}
