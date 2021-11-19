package filter

import (
	"fmt"
	"net/http"
)

type TailFilter struct {
	name string
}

var tailFilter = &TailFilter{name: "tail"}

func init() {
	RegistryFilter(tailFilter)
}

func (t TailFilter) Name() string {
	return t.name
}

func (t TailFilter) Filter(chain *Chain, exchange *Exchange, config interface{}) error {
	exchange.Completed = true
	writeSelf(exchange.Resp, exchange.Req)
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
