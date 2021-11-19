package server

import (
	"errors"
	"github.com/cyejing/shuttle/core/filter"
	"github.com/cyejing/shuttle/pkg/config"
	"github.com/cyejing/shuttle/pkg/log"
	"net/http"
	"strings"
)

type RouteMux struct {
	Routes []config.Route
}

func StartWebServer() {
	c := config.GetConfig()
	routeMux := &RouteMux{Routes: c.Routes}
	if c.Ssl.Enable {
		log.Infof("Start TLS Web Server for addr %s", c.Ssl.Addr)
		err := http.ListenAndServeTLS(c.Ssl.Addr, c.Ssl.Cert, c.Ssl.Key, routeMux)
		if err != nil {
			log.Panic("启动服务失败,请检查证书配置文件", err)
		}
	}
	log.Infof("Start Web Server for addr %s", c.Addr)
	err := http.ListenAndServe(c.Addr, routeMux)
	if err != nil {
		log.Panic("启动服务失败", err)
	}
}

func (r RouteMux) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	route, err := matchRoute(r.Routes, req)
	if err != nil {
		//TODO
		log.Trace("Route not match")
		return
	}
	if route.Loggable {
		log.Debugf("match route %s", route.Id)
	}

	filter.NewChain(resp, req, route).DoFilter()
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
