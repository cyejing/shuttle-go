package server

import (
	"errors"
	"fmt"
	"github.com/cyejing/shuttle/pkg/config"
	"github.com/cyejing/shuttle/pkg/log"
	"net/http"
	"strings"
)

type RouteMux struct {
	Routes []config.Route
}

func StartWebServer() {
	config := config.GetConfig()
	routeMux := &RouteMux{Routes: config.Routes}
	if config.Ssl.Enable {
		log.Infof("Start TLS Web Server for addr %s", config.Ssl.Addr)
		err := http.ListenAndServeTLS(config.Ssl.Addr, config.Ssl.Cert, config.Ssl.Key, routeMux)
		if err != nil {
			log.Panic("启动服务失败,请检查证书配置文件", err)
		}
	}
	log.Infof("Start Web Server for addr %s", config.Addr)
	err := http.ListenAndServe(config.Addr, routeMux)
	if err != nil {
		log.Panic("启动服务失败", err)
	}
}

func (r RouteMux) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	route, err := matchRoute(r.Routes, req)
	if err != nil {
		writeSelf(resp, req)
		fmt.Fprintf(resp, "Route not match\n")
		return
	}
	if route.Loggable {
		log.Debugf("match route %s", route.Id)
	}

	writeSelf(resp, req)
}

func matchRoute(routes []config.Route, req *http.Request) (route config.Route, err error) { //copy Route
	for _, route := range routes {
		if route.Host == req.Host {
			return route, nil
		}
		if route.Path != "" && strings.Index(req.URL.Path, route.Path) == 0 {
			return route, nil
		}
	}
	return config.Route{}, errors.New("route not match")
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
