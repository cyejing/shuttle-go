package codec

import (
	"bufio"
	"github.com/cyejing/shuttle/pkg/config/server"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
	"sync"
)

type Wormhole struct {
	Hash    string
	Name    string
	br      *bufio.Reader
	rwc     net.Conn
	channel chan interface{}
}

func (w *Wormhole) Encode() ([]byte, error) {
	return nil, nil
}
func (w *Wormhole) Decode(r io.Reader) error {
	hash := [56]byte{}
	n, err := r.Read(hash[:])
	if err != nil || n != 56 {
		return utils.BaseErr("failed to read hash", err)
	}
	cc := &ConnectCommand{}
	err = cc.Decode(r)
	if err != nil {
		return utils.BaseErr("decode ConnectCommand fail", err)
	}
	w.Name = cc.name
	return nil
}

var ReqMap sync.Map

func (w *Wormhole) handleReq() error {
	for {
		select {
		case c := <-w.channel:
			if dc, ok := c.(DialCommand); ok {
				ReqMap.Store(dc.reqId, dc.ReqBase)
				bytes, err := dc.Encode()
				if err != nil {
					log.Warnf("encode dial command fail %v", err)
				}
				_, err = w.rwc.Write(bytes)
				if err != nil {
					return utils.BaseErr("handle ReqBase write byte fail", err)
				}
			}
		}
	}
}

func (w *Wormhole) handleResp() error {
	for {
		respC := &RespCommand{}
		err := respC.Decode(w.br)
		if err != nil {
			log.Warnf("handle wormhole name:%s response fail, %v", w.Name, err)
			return utils.BaseErr("handle resp decode response fail", err)
		}
		if r, ok := ReqMap.LoadAndDelete(respC.reqId); ok {
			if loadReq, ok := r.(ReqBase); ok {
				loadReq.respChan <- respC
			}
		}
	}
}

func PeekWormhole(br *bufio.Reader, conn net.Conn) (bool, error) {
	hash, err := br.Peek(56)
	if err != nil {
		return false, utils.BaseErr("peek wormhole fail", err)
	}

	if pw := server.WHPasswords[string(hash)]; pw != nil {
		log.Infof("wormhole %s authenticated as %s", conn.RemoteAddr(), pw.Raw)
		wormhole := &Wormhole{
			Hash:    string(hash),
			br:      br,
			rwc:     conn,
			channel: make(chan interface{}),
		}
		pr := &peekReader{r: br}
		err = wormhole.Decode(pr)
		if err != nil {
			log.Warnf("wormhole proto decode fail %v", err)
			return false, nil
		}

		_, err = br.Discard(pr.i)
		if err != nil {
			log.Warnf("Discard wormhole proto fail %v", err)
			return false, nil
		}

		go func() {
			err := wormhole.handleReq()
			if err != nil {
				log.Warn(err)
			}
		}()

		go func() {
			err := wormhole.handleResp()
			if err != nil {
				log.Warn(err)
			}
		}()

		return true, nil
	}

	return false, nil
}
