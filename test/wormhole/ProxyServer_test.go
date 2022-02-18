package wormhole

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestProxyServer(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:4081")
	if err != nil {
		t.Error(err)
	}

	sendString(t, conn, "hello")
	sendString(t, conn, "look")
	sendString(t, conn, "nice")
}

func sendString(t *testing.T, conn net.Conn, str string) {
	conn.Write([]byte(str))
	r := make([]byte, 1024)
	i, err := conn.Read(r)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, str, string(r[:i]))
}


