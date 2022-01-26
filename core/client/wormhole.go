package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/utils"
	"net"
)

type Wormhole struct {
	Config *client.Config
	Name   string
}

func (w *Wormhole) DialRemote(network, addr string) error {
	var conn net.Conn
	var err error
	if w.Config.SSLEnable {
		conn, err = tls.Dial(network, addr, &tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		conn, err = net.Dial(network, addr)
	}
	if err != nil {
		return utils.BaseErr(fmt.Sprintf("dial remote addr fail %s", addr), err)
	}

	wormhole := &codec.Wormhole{
		Hash:    w.Config.GetHash(),
		Br:      bufio.NewReader(conn),
		Rwc:     conn,
		Channel: make(chan interface{}),
	}

	hashBytes, err := wormhole.Encode()
	if err != nil {
		return utils.BaseErr("wormhole encode fail", err)
	}
	_, err = conn.Write(hashBytes)
	if err != nil {
		return utils.Err(err)
	}
	go func() {
		err := wormhole.HandleCommand()
		if err != nil {
			log.Warn(err)
		}
	}()
	wormhole.Channel <- codec.NewExchangeCommand(w.Name, nil)

	err = wormhole.HandleConn()
	if err != nil {
		return err
	}
	return nil
}
