package client

import (
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/utils"
	"net"
)

type Wormhole struct {
	Config *client.Config
	Name string
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
		return utils.BaseErr(fmt.Sprintf("dial remote addr fail %s",addr),err)
	}

	hash :=w.Config.GetHash()

	conn.Write([]byte(hash))

	cc := codec.NewConnectCommand(w.Name)
	connectByte, err := cc.Encode()
	if err != nil {
		return utils.BaseErr("connect command encode fail", err)
	}
	log.Info(hex.Dump(connectByte))

	_, err = conn.Write(connectByte)
	if err != nil {
		return err
	}

	rc := codec.RespCommand{}
	err = rc.Decode(conn)
	if err != nil {
		return err
	}
	log.Info(rc)
	return nil
}
