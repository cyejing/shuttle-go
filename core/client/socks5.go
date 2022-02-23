package client

import (
	"github.com/cyejing/shuttle/core/codec"
	"github.com/cyejing/shuttle/core/config/client"
	"github.com/cyejing/shuttle/pkg/logger"
	"github.com/cyejing/shuttle/pkg/utils"
	"net"
)
var log = logger.NewLog()


//Socks5Server struct
type Socks5Server struct {
	Config *client.Config
	DialFunc func(config *client.Config, metadata *codec.Metadata) (net.Conn, error)
}

//ListenAndServe listen and serve
func (s *Socks5Server) ListenAndServe(network, addr string) error {
	l, err := net.Listen(network, addr)
	log.Infof("socks5 listen at %s", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Errorln("socks5 accept conn fail ", err)
		}
		go func() {
			defer conn.Close()
			err := s.ServeConn(conn)
			if err != nil {
				log.Errorln("handle socks5 fail ", err)
				return
			}
		}()
	}
}

//ServeConn conn
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

	outbound, err := s.DialFunc(s.Config, socks5.Metadata)
	if err != nil {
		return utils.BaseErrf("socks5 dial remote fail %v", err, outbound)
	}
	defer outbound.Close()

	log.Infof("%s requested connection to %s", outbound.LocalAddr(), socks5.Metadata.String())

	err = socks5.SendReply(codec.SuccessReply)
	if err != nil {
		return utils.BaseErr("socks5 sendReply fail", err)
	}

	return utils.ProxyStream(conn, outbound)
}
