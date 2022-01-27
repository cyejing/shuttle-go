package main

import (
	"flag"
	"github.com/cyejing/shuttle/core/client"
	"github.com/cyejing/shuttle/pkg/codec"
	config "github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/logger"
	"github.com/cyejing/shuttle/pkg/utils"
	"time"
)

var (
	configPath = flag.String("c", "", "config path")
)
var log = logger.NewLog()

func main() {
	flag.Parse()
	c, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}

	err = logger.InitLog(c.LogFile)
	if err != nil {
		panic(err)
	}

	switch c.RunType {
	case "socks":
		socks5 := &client.Socks5Server{
			Config:   c,
			DialFunc: codec.DialTrojan,
		}
		panic(socks5.ListenAndServe("tcp", c.LocalAddr))
	case "wormhole":
		wormhole := &client.Wormhole{
			Config: c,
			Name:   c.Name,
		}
		for {
			func() {
				defer func() {
					if err := recover(); err != nil {

					}
				}()
				err := wormhole.DialRemote("tcp", c.RemoteAddr)
				if err != nil {
					log.Warn(utils.BaseErr("remote conn err", err))
				}
			}()
			time.Sleep(time.Second * 5)
			log.Infof("repeat dial remote %s", c.RemoteAddr)
		}
	}

}
