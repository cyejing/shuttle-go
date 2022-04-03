package channel

import (
	"github.com/cyejing/shuttle/pkg/errors"
	"io"
)

type ProxyChannel struct {
	in  io.ReadWriter
	out io.ReadWriter
}

func NewProxyChannel(in, out io.ReadWriter) *ProxyChannel {
	return &ProxyChannel{
		in:  in,
		out: out,
	}
}

func (p ProxyChannel) Run() error {
	if p.in == nil || p.out == nil {
		return errors.NewErrf("conn is nil, proxy err, %v %v", p.in, p.out)
	}
	ec := make(chan error, 2)
	go func() {
		_, err := io.Copy(p.in, p.out)
		ec <- err
	}()
	go func() {
		_, err := io.Copy(p.out, p.in)
		ec <- err
	}()

	return errors.BaseErr("proxy channel err", <-ec)
}
