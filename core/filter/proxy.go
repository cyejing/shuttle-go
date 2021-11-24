package filter

import (
	"io"
	"net/http"
	"net/url"
)

type proxy struct {
	name string
}

type ProxyConfig struct {
	URI string `asn1:"url"`
}

var client = http.DefaultClient

func init() {
	RegistryFilter(&proxy{name: "proxy"})
}

func (p proxy) Name() string {
	return p.name
}

func (p proxy) Init() {
}

func (p proxy) Filter(exchange *Exchange, c interface{}) error {
	var config ProxyConfig
	if err := mapstruct(c, &config); err != nil {
		return err
	}

	u, err := url.Parse(config.URI + exchange.Req.URL.Path)
	if err != nil {
		return err
	}

	exchange.Req.URL = u
	exchange.Req.RequestURI = ""

	r, err := client.Do(exchange.Req)
	if err != nil {
		return err
	}

	exchange.Resp.WriteHeader(r.StatusCode)
	for k, v := range r.Header {
		exchange.Resp.Header().Set(k, v[0])
	}
	_, err = io.Copy(exchange.Resp, r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return nil
}
