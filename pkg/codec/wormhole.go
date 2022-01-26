package codec

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"github.com/cyejing/shuttle/pkg/config/server"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
	"sync"
)

type Wormhole struct {
	Hash    string
	Name    string
	Br      *bufio.Reader
	Rwc     net.Conn
	Channel chan interface{}
}

func (w *Wormhole) Encode() ([]byte, error) {
	return []byte(w.Hash), nil
}
func (w *Wormhole) Decode(r io.Reader) error {
	hash := [56]byte{}
	n, err := r.Read(hash[:])
	if err != nil || n != 56 {
		return utils.BaseErr("failed to read hash", err)
	}
	return nil
}

var ReqMap sync.Map

func (w *Wormhole) HandleCommand() error {
	buf := bytes.NewBuffer([]byte{})
	for {
		buf.Reset()
		select {
		case c := <-w.Channel:
			buf.WriteByte(byte(ReqType))
			if dc, ok := c.(*ReqBase); ok {
				ReqMap.Store(dc.reqId, dc)
				cBytes, err := dc.Encode()
				if err != nil {
					log.Warnf("encode dial command fail %v", err)
				}
				buf.Write(cBytes)
			} else if ec, ok := c.(*ExchangeCommand); ok {
				//ReqMap.Store(ec.reqId, ec.ReqBase)
				cBytes, err := ec.Encode()
				if err != nil {
					log.Warnf("encode dial command fail %v", err)
				}
				buf.Write(cBytes)
			} else {
				return utils.NewErrf("unknown command %v", c)
			}
			log.Info("write bytes:")
			log.Info(hex.Dump(buf.Bytes()))
			_, err := w.Rwc.Write(buf.Bytes())
			if err != nil {
				return utils.BaseErr("handle ReqBase write byte fail", err)
			}
		}
	}
}

//Command struct
type connType byte

const (
	ReqType connType = iota
	RespType
)

func (w *Wormhole) HandleConn() error {
	for {
		b, err := w.Br.ReadByte()
		if err != nil {
			if err == io.EOF {
				return utils.NewErr("read EOF conn is close")
			}
			return err
		}
		switch connType(b) {
		case ReqType:
			err = w.handleReq()
		case RespType:
			err = w.handleResp()
		default:
			log.Warn("unknown conn type:" + string(b))
		}
		if err != nil {
			return err
		}
	}
}

func (w *Wormhole) handleReq() error {
	ceb, err := w.Br.ReadByte()
	if err != nil {
		return utils.BaseErr("read request command fail", err)
	}
	ce := commandEnum(ceb)
	switch ce {
	case ExchangeCE:
		ec := ExchangeCommand{ReqBase :&ReqBase{commandEnum: ce}}
		err := ec.Decode(w.Br)
		if err != nil {
			return utils.BaseErr("exchange command decode fail", err)
		}
		w.Name = ec.name
		log.Info("exchange name:" + ec.name)
	case DialCE:
		dc := DialCommand{ReqBase :&ReqBase{}}
		err := dc.Decode(w.Br)
		if err != nil {
			return utils.BaseErr("dial command decode fail", err)
		}
		log.Infof("dial decode %v", dc)
	}
	return nil
}

func (w *Wormhole) handleResp() error {
	respC := &RespCommand{}
	err := respC.Decode(w.Br)
	if err != nil {
		return utils.BaseErr("handle resp decode response command fail", err)
	}
	if r, ok := ReqMap.LoadAndDelete(respC.reqId); ok {
		if loadReq, ok := r.(ReqBase); ok {
			loadReq.respChan <- respC
		}
	}
	return nil
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
			Br:      br,
			Rwc:     conn,
			Channel: make(chan interface{}),
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
			err := wormhole.HandleCommand()
			if err != nil {
				log.Warn(err)
			}
		}()

		err := wormhole.HandleConn()
		if err != nil {
			log.Warn(err)
		}
		return true, nil
	}

	return false, nil
}
