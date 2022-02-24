package filter

import (
	"context"
	"errors"
	"fmt"
	config "github.com/cyejing/shuttle/core/config/server"
	"github.com/cyejing/shuttle/pkg/logger"
	"github.com/goinggo/mapstructure"
	"net/http"
	"regexp"
	"sort"
	"sync"
)

var log = logger.NewLog()

// registry filter for chain
var registryFilters = map[string]Filter{}

// RegistryFilter register filter
func RegistryFilter(filter Filter) {
	registryFilters[filter.Name()] = filter
}

// Filter interface
type Filter interface {
	Init(mux *RouteMux)
	Name() string
	Filter(exchange *Exchange, c interface{}) error
}

// Exchange struct
type Exchange struct {
	Resp      http.ResponseWriter
	Req       *http.Request
	Ctx       context.Context
	Attr      map[string]interface{}
	completed bool
	Written   sync.Once
}

func (e *Exchange) Completed() {
	e.completed = true
}

//RouteMux struct
type RouteMux struct {
	Routes []config.Route
}

//NewRouteMux new route mux
func NewRouteMux(c *config.Config) *RouteMux {
	routeMux := &RouteMux{Routes: c.Gateway.Routes}

	for _, filter := range registryFilters {
		filter.Init(routeMux)
	}

	sort.Slice(routeMux.Routes, func(i, j int) bool {
		return routeMux.Routes[i].Order > routeMux.Routes[j].Order
	})

	return routeMux
}

func (r *RouteMux) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	route, err := r.matchRoute(req)
	if err != nil {
		log.Debug(err)
		write404(resp)
		return
	}
	if route.Loggable {
		log.Debugf("match route %s", route.ID)
	}

	NewChain(resp, req, route).DoFilter()
}

func (r *RouteMux) matchRoute(req *http.Request) (route config.Route, err error) { //copy Route
	for _, route := range r.Routes {
		if ok, _ := regexp.MatchString(route.Host, req.Host); route.Host != "" && ok {
			return route, nil
		}
		if ok, _ := regexp.MatchString(route.Path, req.URL.Path); route.Path != "" && ok {
			return route, nil
		}
	}
	return config.Route{}, errors.New("route not match")
}

// Chain struct
type Chain struct {
	Filters  []Filter
	Index    int
	Exchange *Exchange
	Route    config.Route
}

// NewChain new chain
func NewChain(resp http.ResponseWriter, req *http.Request, route config.Route) *Chain {
	var filters = make([]Filter, len(route.Filters))
	for i, filter := range route.Filters {
		filters[i] = registryFilters[filter.Name]
	}
	return &Chain{
		Filters: filters,
		Index:   0,
		Exchange: &Exchange{
			Resp: resp,
			Req:  req,
			Ctx:  context.Background(),
			Attr: make(map[string]interface{}),
		},
		Route: route,
	}
}

//DoFilter run filter
func (c *Chain) DoFilter() {
	for _, f := range c.Filters {
		fc, err := c.Route.GetFilter(f.Name())
		if err != nil {
			log.Warn(err)
			continue
		}
		err = f.Filter(c.Exchange, fc.Params)

		if err != nil {
			if re, ok := err.(*RespErr); ok {
				log.Warn(re)
				re.Html(c.Exchange.Resp)
			} else {
				log.Error(err)
				write500(c.Exchange.Resp)
			}
			return
		}

		if c.Exchange.completed {
			c.complete()
			return
		}
	}

	if !c.Exchange.completed {
		writeStatusCode(c.Exchange.Resp, 502)
	}
}

func (c *Chain) complete() {

}

type RespErr struct {
	code int
	msg  string
}

func NewRespErr(code int, msg string) *RespErr {
	return &RespErr{
		code: code,
		msg:  msg,
	}
}

func (r *RespErr) Error() string {
	return fmt.Sprintf("%d %s : %s :", r.code, http.StatusText(r.code), r.msg)
}

func (r *RespErr) Html(resp http.ResponseWriter) {
	fmt.Fprintf(resp, html, r.code, http.StatusText(r.code), r.code, http.StatusText(r.code))
}

//more func
func mapstruct(c interface{}, config interface{}) error {
	return mapstructure.Decode(c, config)
}

var (
	html     = "<html>\n<head><title>%d %s</title></head>\n<body>\n<center><h1>%d %s</h1></center>\n<hr><center>nginx</center>\n</body>\n</html>"
	respHtml = "<html>\n<head><title>%d %s</title></head>\n<body>\n<center><h1>%d %s</h1></center>\n<center><p>%s</p></center>\n<hr><center>nginx</center>\n</body>\n</html>"
)

func writeStatusCode(resp http.ResponseWriter, code int) {
	fmt.Fprintf(resp, html, code, http.StatusText(code), code, http.StatusText(code))
}

func write404(resp http.ResponseWriter) {
	writeStatusCode(resp, 404)
}

func write500(resp http.ResponseWriter) {
	writeStatusCode(resp, 500)
}
