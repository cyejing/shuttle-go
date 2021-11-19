package filter

import (
	"github.com/goinggo/mapstructure"
	"io"
	"net/http"
	"net/url"
)

type ProxyFilter struct {
	name string
}

type Config struct {
	Uri string
}

var proxyFilter = &ProxyFilter{name: "proxy"}
var client = http.DefaultClient

func init() {
	RegistryFilter(proxyFilter)
}

func (p ProxyFilter) Name() string {
	return p.name
}

func (p ProxyFilter) Filter(chain *Chain, exchange *Exchange, c interface{}) error {
	var config Config
	if err := mapstructure.Decode(c, &config); err != nil {
		return err
	}
	url, err := url.Parse(config.Uri + exchange.Req.RequestURI)
	if err != nil {
		return err
	}

	exchange.Req.URL = url
	exchange.Req.RequestURI = ""

	r, err := client.Do(exchange.Req)
	if err != nil {
		return err
	}

	exchange.Resp.WriteHeader(r.StatusCode)
	for k, v := range r.Header {
		exchange.Resp.Header().Set(k, v[0])
	}
	io.Copy(exchange.Resp, r.Body)

	chain.DoFilter()
	return nil
}
