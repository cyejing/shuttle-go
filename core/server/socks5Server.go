package server

import (
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/utils"
	"net"
)

type Socks5Server struct {
	DialFunc func(metadata *codec.Metadata) (net.Conn, error)
}

func (s *Socks5Server) ListenAndServe(network, addr string) error {
	l, err := net.Listen(network, addr)
	log.Infof("socks5 listen at %s", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error("socks5 accept conn fail ", err)
		}
		go func() {
			defer conn.Close()
			err := s.ServeConn(conn)
			if err != nil {
				log.Error("handle socks5 fail ", err)
				return
			}
		}()
	}
}

func (s *Socks5Server) ServeConn(conn net.Conn) (err error) {
	socks5 := codec.Socks5{Conn: conn}

	err = socks5.HandleHandshake()
	if err != nil {
		return utils.BaseErr("socks5 HandleHandshake fail", err)
	}
	err = socks5.LSTRequest()
	if err != nil {
		return utils.BaseErr("socks5 LSTRequest fail", err)
	}

	outbound, err := s.DialFunc(socks5.Metadata)
	if err != nil {
		return utils.BaseErrf("socks5 dial remote fail %v", err, outbound.RemoteAddr())
	}
	defer outbound.Close()

	log.Infof("%s requested connection to %s", outbound.LocalAddr(), socks5.Metadata.String())

	err = socks5.SendReply(codec.SuccessReply)
	if err != nil {
		return utils.BaseErr("socks5 sendReply fail", err)
	}

	return utils.ProxyStream(conn, outbound)
}
