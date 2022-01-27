package codec

import (
	"bufio"
	"encoding/binary"
	"github.com/cyejing/shuttle/pkg/logger"
	"io"
)

var log = logger.NewLog()

const (
	maxPacketSize = 1024 * 8
)

//Codec interface
type Codec interface {
	Encode() ([]byte, error)
	Decode(reader io.Reader) error
}

type PeekReader struct {
	R *bufio.Reader
	I int
}

func (p *PeekReader) Read(b []byte) (n int, err error) {
	peek, err := p.R.Peek(p.I + len(b))
	if err != nil {
		return 0, err
	}
	ci := copy(b, peek[p.I:])
	p.I += ci
	return ci, nil
}

func ReadByte(r io.Reader) (byte, error) {
	bytes := [1]byte{}
	_, err := r.Read(bytes[:])
	if err != nil {
		return 0, err
	}
	return bytes[0], nil
}

func ReadUint32(r io.Reader) (uint32, error) {
	bytes := [4]byte{}
	_, err := io.ReadFull(r, bytes[:])
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(bytes[:]), nil
}
