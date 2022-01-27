package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/operate"
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

	wormhole := &operate.Wormhole{
		Name: w.Name,
		Hash: w.Config.GetHash(),
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

	err = operate.NewDispatcher(wormhole).Exchange()
	if err != nil {
		return err
	}
	return nil
}
