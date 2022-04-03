package channel

import (
	"bufio"
	"github.com/cyejing/shuttle/core/config"
	"github.com/cyejing/shuttle/pkg/errors"
	"net"
)

type PeekChannel struct {
	config config.ServerConfig
	raw    net.Conn
	proxy  net.Conn
	buf    *bufio.Reader
}

func NewPeekChannel(raw net.Conn, c config.ServerConfig) *PeekChannel {
	proxy, err := net.Dial("tcp", c.Trojan.Addr)
	if err != nil {
		log.Warn(errors.BaseErr("dial trojan proxy addr err", err))
	}
	return &PeekChannel{
		config: c,
		raw:    raw,
		proxy:  proxy,
		buf:    bufio.NewReader(raw),
	}
}

func (p *PeekChannel) Run() error {
	hash, err := p.buf.Peek(56)
	if err != nil {
		return errors.BaseErr("peek channel err", err)
	}
	if p.isTrojan(hash) {
		return NewTrojanChannel().Run()
	}
	if p.isWormhole(hash) {
		return NewWormholeChannel(hash).Run()
	}

	return NewProxyChannel(p.raw, p.proxy).Run()
}

func (p *PeekChannel) isTrojan(hash []byte) bool {
	pw := p.config.Trojan.PasswordMap[string(hash)]
	return pw != nil
}

func (p *PeekChannel) isWormhole(hash []byte) bool {
	pw := p.config.Wormhole.PasswordMap[string(hash)]
	return pw != nil
}
