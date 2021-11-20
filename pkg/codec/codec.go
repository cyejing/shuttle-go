package codec

import "io"

const (
	maxPacketSize = 1024 * 8
)

type Codec interface {
	Encode() ([]byte, error)
	Decode(reader io.Reader) error
}
