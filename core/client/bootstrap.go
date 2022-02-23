package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/cyejing/shuttle/core/codec"
	config "github.com/cyejing/shuttle/core/config/client"
	operate2 "github.com/cyejing/shuttle/core/operate"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
	"time"
)

func Run(c *config.Config) {
	switch c.RunType {
	case "socks":
		runSocks(c)
	case "wormhole":
		loopRunWormhole(c)
	}
	log.Infof("client exit")
}

func runSocks(c *config.Config) {
	socks5 := &Socks5Server{
		Config:   c,
		DialFunc: codec.DialTrojan,
	}
	panic(socks5.ListenAndServe("tcp", c.SockAddr))
}

func loopRunWormhole(c *config.Config) {
	for {
		err := dialRemote(c)
		if err != nil {
			if err == io.EOF {
				log.Info("remote conn close, reconnect later")
			} else {
				log.Error(utils.BaseErr("remote conn err", err))
			}
		}

		time.Sleep(time.Second * 5)
		log.Infof("repeat dial remote %s", c.RemoteAddr)
	}
}

func dialRemote(c *config.Config) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error(utils.NewErrf("run wormhole dial remote catch err %v", err))
		}
	}()

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

	wormhole := &operate2.Wormhole{
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

	err = operate2.NewCliDispatcher(wormhole, c.Name).Connect()
	if err != nil {
		return err
	}
	return nil
}
