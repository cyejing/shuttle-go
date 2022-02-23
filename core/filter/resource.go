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

// ResourceConfig struct
type ResourceConfig struct {
	Root string
}

func (r resource) Init(mux *RouteMux) {

}

func (r resource) Name() string {
	return r.name
}

var indexHTML = []string{"index.html", "index.htm", "/index.html", "/index.htm"}

func (r resource) Filter(exchange *Exchange, c interface{}) error {
	var config ResourceConfig
	if err := mapstruct(c, &config); err != nil {
		return err
	}

	path := config.Root + exchange.Req.URL.Path[1:]
	paths := make([]string, len(indexHTML)+1)
	paths[0] = path
	for i, s := range indexHTML {
		paths[i+1] = path + s
	}

	var file *os.File
	var err error
	for _, p := range paths {
		file, err = os.OpenFile(p, os.O_RDONLY, 0)
		if err != nil {
			continue
		}
		var stat os.FileInfo
		stat, err = file.Stat()

		if err == nil && !stat.IsDir() {
			break
		}
	}
	if err != nil {
		if os.IsNotExist(err) {
			write404(exchange.Resp)
			exchange.Completed()
		}
		return nil
	}

	_, err = io.Copy(exchange.Resp, file) // auto sendfile, good job
	if err != nil {
		return err
	}

	defer func() {
		exchange.Completed()
		if file != nil {
			file.Close()
		}
	}()
	return nil
}

