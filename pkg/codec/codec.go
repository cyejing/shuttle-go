package codec

import (
	"github.com/cyejing/shuttle/pkg/logger"
	"io"
)

var log = logger.NewLog()

const (
	maxPacketSize = 1024 * 8
)

type Codec interface {
	Encode() ([]byte, error)
	Decode(reader io.Reader) error
}
