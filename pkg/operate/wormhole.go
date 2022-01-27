package operate

import (
	"bufio"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/config/server"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
)

type Wormhole struct {
	Hash    string
	Name    string
	Br      *bufio.Reader
	Rwc     net.Conn
}

func (w *Wormhole) Encode() ([]byte, error) {
	return []byte(w.Hash), nil
}
func (w *Wormhole) Decode(r io.Reader) error {
	hash := [56]byte{}
	n, err := r.Read(hash[:])
	if err != nil || n != 56 {
		return utils.BaseErr("failed to read hash", err)
	}
	return nil
}

func PeekWormhole(br *bufio.Reader, conn net.Conn) (bool, error) {
	hash, err := br.Peek(56)
	if err != nil {
		return false, utils.BaseErr("peek wormhole fail", err)
	}

	if pw := server.WHPasswords[string(hash)]; pw != nil {
		log.Infof("wormhole %s authenticated as %s", conn.RemoteAddr(), pw.Raw)
		wormhole := &Wormhole{
			Hash:    string(hash),
			Br:      br,
			Rwc:     conn,
		}
		pr := &codec.PeekReader{R: br}
		err = wormhole.Decode(pr)
		if err != nil {
			log.Warnf("wormhole proto decode fail %v", err)
			return false, nil
		}

		_, err = br.Discard(pr.I)
		if err != nil {
			log.Warnf("Discard wormhole proto fail %v", err)
			return false, nil
		}

		err := NewDispatcher(wormhole).Run()
		if err != nil {
			log.Warn(err)
		}

		return true, nil
	}

	return false, nil
}
