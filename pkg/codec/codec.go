package codec

import (
	"bufio"
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


type peekReader struct {
	r *bufio.Reader
	i int
}

func (p *peekReader) Read(b []byte) (n int, err error) {
	peek, err := p.r.Peek(p.i + len(b))
	if err != nil {
		return 0, err
	}
	ci := copy(b, peek[p.i:])
	p.i += ci
	return ci, nil
}
