package filter

import (
	"bufio"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/common"
	"github.com/cyejing/shuttle/pkg/config"
	"github.com/cyejing/shuttle/pkg/log"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
)

type socks struct {
	name string
}

type Password struct {
	raw  string
	hash string
}

var (
	passwords = make(map[string]*Password)
)

func init() {
	RegistryFilter(&socks{"socks"})
}

func (t socks) Init() {
	for _, raw := range config.GetConfig().Passwords {
		hash := utils.SHA224String(raw)
		passwords[hash] = &Password{
			raw:  raw,
			hash: hash,
		}
	}
}

func (t socks) Name() string {
	return t.name
}

func (t socks) Filter(exchange *Exchange, config interface{}) error {
	bufBody := bufio.NewReader(exchange.Req.Body)
	//exchange.Req.Body.Close()
	exchange.Req.Body = io.NopCloser(bufBody)

	peek, err := bufBody.Peek(56)
	if err != nil {
		log.Error("socks peek err", err)
		return nil
	}
	if passwords[string(peek)] != nil {
		socks := new(codec.Socks)
		socks.Decode(bufBody)
		if inbound, ok := exchange.Req.Context().Value(common.ConnContextKey).(net.Conn); ok {
			outbound, err := net.Dial("tcp", socks.Metadata.Address.String())
			if err != nil {
				log.Error("socks dial addr err %v %v", socks.Metadata.Address.String(), err)
				return nil
			}

			// Start proxying
			errCh := make(chan error, 2)
			go connProxy(outbound, inbound, errCh)
			go connProxy(inbound, outbound, errCh)
			// Wait
			for i := 0; i < 2; i++ {
				e := <-errCh
				if e != nil {
					// return from this function closes target (and conn).
					return e
				}
			}
		}
	} else {
		log.Warnf("socks password auth fail")
	}

	return nil
}

func PeekTrojanProto(b []byte) bool {
	hash := string(b)
	return passwords[hash] != nil
}

type closeWriter interface {
	CloseWrite() error
}

func connProxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	if tcpConn, ok := dst.(closeWriter); ok {
		tcpConn.CloseWrite()
	}
	errCh <- err
}
