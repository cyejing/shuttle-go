package pkg

import (
	"bytes"
	"errors"
	"io"
	"net"
)

type Pipeline struct {
	conn     *net.TCPConn
	handlers []Handler
}

type Handler interface {
	Inbound(pipe Pipeline, reader io.Reader) (err error)
	Outbound(msg interface{}) (r interface{}, err error)
}

type firstHandler struct {
	conn *net.TCPConn
}

type lastHandler struct{}

func NewTCP(conn *net.TCPConn, hs []Handler) *Pipeline {
	newhs := initPipe(conn, hs)
	return &Pipeline{
		conn:     conn,
		handlers: newhs,
	}
}

func initPipe(conn *net.TCPConn, hs []Handler) []Handler {
	newhs := append([]Handler{}, firstHandler{conn: conn})
	for _, h := range hs {
		newhs = append(newhs, h)
	}
	newhs = append(newhs, lastHandler{})
	return newhs
}

func (p *Pipeline) Handler() (err error) {
	for _, handler := range p.handlers {
		err := handler.Inbound(*p, p.conn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) Write(o interface{}) (err error) {
	nr := o
	for _, handler := range p.handlers {
		r, err := handler.Outbound(nr)
		if err != nil {
			return err
		}
		nr = r
	}
	return nil
}

func (h firstHandler) Inbound(pipe Pipeline, reader io.Reader) (err error) {
	return nil
}
func (h firstHandler) Outbound(msg interface{}) (r interface{}, err error) {
	if msg == nil {
		return nil, errors.New("outbound msg nil")
	}

	if b, ok := msg.([]byte); ok {
		_, err := h.conn.Write(b)
		if err != nil {
			return nil, err
		}
	}
	if buf, ok := msg.(*bytes.Buffer); ok {
		_, err := h.conn.Write(buf.Bytes())
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (h lastHandler) Inbound(pipe Pipeline, reader io.Reader) (err error) {
	return nil
}
func (h lastHandler) Outbound(msg interface{}) (r interface{}, err error) {
	return nil, nil
}
