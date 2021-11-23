package server

import (
	"crypto/tls"
	"fmt"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/utils"
	"net"
)

type Socks5Server struct {
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
			log.Error("socks5 accept conn fail |", err)
		}
		go func() {
			defer conn.Close()
			err := s.ServeConn(conn)
			if err != nil {
				log.Error("handle socks5 fail |", err)
				return
			}
		}()
	}
}

func (s *Socks5Server) ServeConn(conn net.Conn) (err error) {
	config := client.GetConfig()
	socks5 := codec.Socks5{Conn: conn}

	err = socks5.HandleHandshake()
	if err != nil {
		return utils.NewError("socks5 HandleHandshake fail").Base(err)
	}
	err = socks5.LSTRequest()
	if err != nil {
		return utils.NewError("socks5 LSTRequest fail").Base(err)
	}

	outbound, err := tls.Dial("tcp", config.RemoteAddr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return utils.NewError(fmt.Sprintf("socks5 dial remote fail %v", config.RemoteAddr)).Base(err)
	}
	defer outbound.Close()

	log.Infof("%s requested connection to %s", outbound.LocalAddr(), socks5.Metadata.String())

	err = socks5.SendReply(codec.SuccessReply)
	if err != nil {
		return utils.NewError("socks5 sendReply fail").Base(err)
	}
	err = sendTrojan(outbound, socks5.Metadata.Address)
	if err != nil {
		return utils.NewError("socks5 sendTrojan fail").Base(err)
	}

	return utils.ProxyStream(conn, outbound)
}

func sendTrojan(outbound net.Conn, address *codec.Address) error {
	c := client.GetConfig()

	socks := &codec.Trojan{
		Hash: utils.SHA224String(c.Password),
		Metadata: &codec.Metadata{
			Command: codec.Connect,
			Address: address,
		},
	}
	encode, err := socks.Encode()
	if err != nil {
		return err
	}

	_, err = outbound.Write(encode)
	if err != nil {
		return err
	}
	return nil
}
