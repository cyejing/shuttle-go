package handler

import (
	"bufio"
	"github.com/cyejing/shuttle/pkg"
	"io"
)

type Echo struct{}

func (e Echo) Inbound(pipe pkg.Pipeline, reader io.Reader) (err error) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		pipe.Write(scanner.Bytes())
	}
	return nil
}

func (e Echo) Outbound(msg interface{}) (r interface{}, err error) {
	return nil, nil
}
