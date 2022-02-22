package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/cyejing/shuttle/pkg/codec"
	config "github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/operate"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
	"time"
)

func Run(c *config.Config) {
	switch c.RunType {
	case "socks":
		socks5 := &Socks5Server{
			Config:   c,
			DialFunc: codec.DialTrojan,
		}
		panic(socks5.ListenAndServe("tcp", c.SockAddr))
	case "wormhole":
		for {
			func() {
				defer func() {
					if err := recover(); err != nil {

					}
				}()
				err := DialRemote(c)
				if err != nil {
					if err == io.EOF {
						log.Info("remote conn close, reconnect later")
					}else{
						log.Error(utils.BaseErr("remote conn err", err))
					}
				}
			}()
			time.Sleep(time.Second * 5)
			log.Infof("repeat dial remote %s", c.RemoteAddr)
		}
	}
	log.Infof("client exit")
}

func DialRemote(c *config.Config) error {
	var conn net.Conn
	var err error
	network := "tcp"
	addr := c.RemoteAddr
	if c.SSLEnable {
		conn, err = tls.Dial(network, addr, &tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		conn, err = net.Dial(network, addr)
	}
	if err != nil {
		return utils.BaseErr(fmt.Sprintf("dial remote addr fail %s", addr), err)
	}

	wormhole := &operate.Wormhole{
		Hash: c.GetHash(),
		Br:   bufio.NewReader(conn),
		Rwc:  conn,
	}

	hashBytes, err := wormhole.Encode()
	if err != nil {
		return utils.BaseErr("wormhole encode fail", err)
	}
	_, err = conn.Write(hashBytes)
	if err != nil {
		return utils.Err(err)
	}

	err = operate.NewCliDispatcher(wormhole, c.Name).Connect()
	if err != nil {
		return err
	}
	return nil
}
