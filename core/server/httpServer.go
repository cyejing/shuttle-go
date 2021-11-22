package server

import (
	"errors"
	"github.com/cyejing/shuttle/core/filter"
	config "github.com/cyejing/shuttle/pkg/config/server"
	"github.com/cyejing/shuttle/pkg/log"
	"net/http"
	"sort"
	"strings"
)

type RouteMux struct {
	Routes []config.Route
}

func NewRouteMux() *RouteMux {
	c := config.GetConfig()

	routeMux := &RouteMux{Routes: c.Routes}

	sort.Slice(routeMux.Routes, func(i, j int) bool {
		return routeMux.Routes[i].Order > routeMux.Routes[j].Order
	})

	filter.Init()

	return routeMux
}

func (r RouteMux) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	route, err := matchRoute(r.Routes, req)
	if err != nil {
		//TODO
		log.L.Trace(err)
		return
	}
	if route.Loggable {
		log.L.Debugf("match route %s", route.Id)
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
