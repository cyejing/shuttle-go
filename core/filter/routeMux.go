package filter

import (
	"errors"
	config "github.com/cyejing/shuttle/pkg/config/server"
	"net/http"
	"regexp"
	"sort"
)

//RouteMux struct
type RouteMux struct {
	Routes []config.Route
}

//NewRouteMux new route mux
func NewRouteMux(c *config.Config) *RouteMux {
	routeMux := &RouteMux{Routes: c.Gateway.Routes}

	Init(routeMux)

	sort.Slice(routeMux.Routes, func(i, j int) bool {
		return routeMux.Routes[i].Order > routeMux.Routes[j].Order
	})

	return routeMux
}

func (r RouteMux) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	route, err := matchRoute(r.Routes, req)
	if err != nil {
		log.Trace(err)
		return
	}
	if route.Loggable {
		log.Debugf("match route %s", route.ID)
	}

	NewChain(resp, req, route).DoFilter()
}

func matchRoute(routes []config.Route, req *http.Request) (route config.Route, err error) { //copy Route
	for _, route := range routes {
		if ok, _ := regexp.MatchString(route.Host, req.Host); route.Host != "" && ok {
			return route, nil
		}
		if ok, _ := regexp.MatchString(route.Path, req.URL.Path); route.Path != "" && ok {
			return route, nil
		}
	}
	return config.Route{}, errors.New("route not match")
}
