package server

import (
	"fmt"
	"github.com/cyejing/shuttle/pkg/config"
	"github.com/cyejing/shuttle/pkg/log"
	"net/http"
)

type RouteMux struct {
	Routes []config.Route
}

func StartWebServer(addr string) {
	config := config.GetConfig()
	routeMux := &RouteMux{Routes: config.Routes}

	if config.Ssl.Enable {
		log.Infof("Start TLS Web Server for addr %s", addr)
		err := http.ListenAndServeTLS(addr, config.Ssl.Cert, config.Ssl.Key, routeMux)
		if err != nil {
			log.Panic("启动服务失败,请检查证书配置文件", err)
		}
	} else {
		log.Infof("Start Web Server for addr %s", addr)
		err := http.ListenAndServe(addr, routeMux)
		if err != nil {
			log.Panic("启动服务失败", err)
		}
	}
}

func (r RouteMux) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
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
