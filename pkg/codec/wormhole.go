package codec

import (
	"bufio"
	"net"
)

type Wormhole struct {
	Hash string
}

func PeekWormhole(reader *bufio.Reader, conn net.Conn) (bool, error) {
	return false, nil
}
