package server

import (
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/log"
	"io"
	"net"
	"os"
)

type Socks5Server struct {
}

func (s *Socks5Server) ListenAndServe(network, addr string) error {
	l, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error("socks5 accept conn err", err)
			return err
		}
		go func() {
			defer conn.Close()
			err := s.ServeConn(conn)
			if err != nil {
				log.Error("handle socks5 err", err)
				return
			}
		}()
	}
	return nil
}

func (s *Socks5Server) ServeConn(conn net.Conn) (err error) {
	log.Debugf("accept socks5 conn", conn)

	config := client.GetConfig()

	socks5 := codec.Socks5{Conn: conn}

	err = socks5.HandleHandshake()

	err = socks5.LSTRequest()

	outbound, err := socks5.DialRemote("tcp", config.RemoteAddr)
	if err != nil {
		log.Error(conn.RemoteAddr(), err)
		return
	}
	defer outbound.Close()

	err = socks5.SendReply(codec.SuccessReply)
	if err != nil {
		return err
	}
	lr := &logReader{r: outbound, w: outbound}
	// Start proxying
	errCh := make(chan error, 2)
	go connProxy(lr, conn, errCh)
	go connProxy(conn, lr, errCh)
	// Wait
	for i := 0; i < 2; i++ {
		e := <-errCh
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

type logReader struct {
	r io.Reader
	w io.Writer
}

func (l *logReader) Write(p []byte) (n int, err error) {
	var f, _ = os.OpenFile("Write.file", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	f.Write(p)
	defer f.Close()
	return l.w.Write(p)
}

func (l *logReader) Read(p []byte) (n int, err error) {
	var f, _ = os.OpenFile("Read.file", os.O_WRONLY|os.O_CREATE, 0666)
	f.Write(p)
	defer f.Close()
	return l.r.Read(p)
}

func connProxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	if tcpConn, ok := dst.(closeWriter); ok {
		tcpConn.CloseWrite()
	}
	errCh <- err
}
