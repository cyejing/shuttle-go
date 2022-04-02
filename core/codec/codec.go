package codec

import (
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

func DecodeUint32(uintByte []byte) uint32 {
	return binary.BigEndian.Uint32(uintByte)
}

func EncodeUint32(i uint32) []byte {
	uintByte := [4]byte{}
	binary.BigEndian.PutUint32(uintByte[:], i)
	return uintByte[:]
}
