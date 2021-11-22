package utils

import (
	"io"
)

func ProxyStream(r io.ReadWriter, w io.ReadWriter) error {
	ec := make(chan error, 2)
	go proxyStream(r, w, ec)
	go proxyStream(w, r, ec)
	for i := 0; i < 2; i++ {
		e := <-ec
		if e != nil {
			// return from this function closes target (and conn).
			return e
		}
	}
	return nil
}

type closeWriter interface {
	CloseWrite() error
}

func proxyStream(dst io.Writer, src io.Reader, errCh chan error) {
	//buf := bufio.NewReader(src)
	//peek, _ := buf.Peek(64)
	//log.Debugf("read bytes:\n%s", hex.Dump(peek))
	_, err := io.Copy(dst, src)
	if tcpConn, ok := dst.(closeWriter); ok {
		err = tcpConn.CloseWrite()
	}
	errCh <- err
}
