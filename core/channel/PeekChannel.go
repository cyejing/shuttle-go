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
	buf    bufio.Reader
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

	return NewProxyChannel(p.raw,).Run()
}


func (p *PeekChannel) isTrojan(hash []byte) bool {
	pw := p.config.Trojan.PasswordMap[string(hash)]
	return pw != nil
}

func (p *PeekChannel) isWormhole(hash []byte) bool {
	pw := p.config.Wormhole.PasswordMap[string(hash)]
	return pw != nil
}

type PeekReader struct {
	R *bufio.Reader
	I int
}

func (p *PeekReader) Read(b []byte) (n int, err error) {
	peek, err := p.R.Peek(p.I + len(b))
	if err != nil {
		return 0, err
	}
	ci := copy(b, peek[p.I:])
	p.I += ci
	return ci, nil
}
