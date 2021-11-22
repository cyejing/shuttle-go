package filter

import (
	"bufio"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/common"
	"github.com/cyejing/shuttle/pkg/log"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
)

type socks struct {
	name string
}

func init() {
	RegistryFilter(&socks{"socks"})
}

func (t socks) Init() {

}

func (t socks) Name() string {
	return t.name
}

func (t socks) Filter(exchange *Exchange, config interface{}) error {
	bufBody := bufio.NewReader(exchange.Req.Body)
	exchange.Req.Body = io.NopCloser(bufBody)

	peek, err := bufBody.Peek(56)
	if err != nil {
		log.L.Error("socks peek err", err)
		return nil
	}
	if codec.ExitHash(peek) {
		trojan := new(codec.Trojan)
		trojan.Decode(bufBody)
		if inbound, ok := exchange.Req.Context().Value(common.ConnContextKey).(net.Conn); ok {
			outbound, err := net.Dial("tcp", trojan.Metadata.Address.String())
			if err != nil {
				log.L.Error("socks dial addr err %v %v", trojan.Metadata.Address.String(), err)
				return nil
			}
			defer outbound.Close()

			utils.ProxyStream(inbound, outbound)
		}
	} else {
		log.L.Warnf("socks password auth fail")
	}

	return nil
}
