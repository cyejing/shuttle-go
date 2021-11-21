package filter

import (
	"bytes"
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

var indexHtml = []string{"index.html", "index.htm", "/index.html", "/index.htm"}
var html404 = "<html>\n<head><title>404 Not Found</title></head>\n<body>\n<center><h1>404 Not Found</h1></center>\n<hr>\n</body>\n</html>"

func (r resource) Filter(exchange *Exchange, c interface{}) error {
	var config ResourceConfig
	if err := mapstruct(c, &config); err != nil {
		return err
	}

	path := config.Root + exchange.Req.URL.Path[1:]
	paths := make([]string, len(indexHtml)+1)
	paths[0] = path
	for i, s := range indexHtml {
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
			buf404 := bytes.NewBufferString(html404)
			_, err = io.Copy(exchange.Resp, buf404)
		}
		return err
	}

	_, err = io.Copy(exchange.Resp, file) // auto sendfile, good job
	if err != nil {
		return err
	}

	return nil
}
