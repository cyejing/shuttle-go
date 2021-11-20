package filter

import (
	"io"
	"os"
)

type resource struct {
	name string
}

func init() {
	RegistryFilter(&resource{name: "resource"})
}

type ResourceConfig struct {
	Root string
}

func (r resource) Init() {

}

func (r resource) Name() string {
	return r.name
}

func (r resource) Filter(exchange *Exchange, c interface{}) error {
	var config ResourceConfig
	if err := mapstruct(c, &config); err != nil {
		return err
	}

	path := config.Root + exchange.Req.URL.Path[1:]
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	_, err = io.Copy(exchange.Resp, file) // auto sendfile, good job
	if err != nil {
		return err
	}

	return nil
}
