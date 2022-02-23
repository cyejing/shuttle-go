package filter

import (
	"fmt"
	"net/http"
)

type tail struct {
	name string
}

func init() {
	RegistryFilter(&tail{name: "tail"})
}

func (t tail) Init(mux *RouteMux) {
}

func (t tail) Name() string {
	return t.name
}

func (t tail) Filter(exchange *Exchange, config interface{}) error {
	log.Infof("filter tail")
	writeSelf(exchange.Resp, exchange.Req)
	exchange.Completed()
	return nil
}

func writeSelf(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(resp, "%s %s %s\n", req.Method, req.URL, req.Proto)
	for k, v := range req.Header {
		fmt.Fprintf(resp, "Header[%q] = %q\n", k, v)
	}
	fmt.Fprintf(resp, "Host = %q\n", req.Host)
	fmt.Fprintf(resp, "RemoteAddr = %q\n", req.RemoteAddr)
	if err := req.ParseForm(); err != nil {
	}
	for k, v := range req.Form {
		fmt.Fprintf(resp, "Form[%q] = %q\n", k, v)
	}
}
