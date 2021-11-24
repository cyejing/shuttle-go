package utils

import (
	"io"
)

// ProxyStreamBuf swap byte
func ProxyStreamBuf(r1 io.Reader, w1 io.Writer, r2 io.Reader, w2 io.Writer) error {
	ec := make(chan error, 2)
	go proxyStream(w1, r2, ec)
	go proxyStream(w2, r1, ec)
	for i := 0; i < 2; i++ {
		e := <-ec
		if e != nil {
			// return from this function closes target (and conn).
			return e
		}
	}
	return nil
}

// ProxyStream swap byte
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
