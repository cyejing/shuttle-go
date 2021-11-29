package filter

import (
	"context"
	config "github.com/cyejing/shuttle/pkg/config/server"
	"github.com/cyejing/shuttle/pkg/logger"
	"github.com/goinggo/mapstructure"
	"net/http"
	"sync"
)

var log = logger.NewLog()

// Filter interface
type Filter interface {
	Init()
	Name() string
	Filter(exchange *Exchange, c interface{}) error
}

// Exchange struct
type Exchange struct {
	Resp      http.ResponseWriter
	Req       *http.Request
	Ctx       context.Context
	Attr      map[string]interface{}
	Err       error
	Completed bool
	Written   sync.Once
}

func (e *Exchange) Error(err error) {
	e.Completed = true
	e.Err = err
}

// Chain struct
type Chain struct {
	Filters  []Filter
	Index    int
	Exchange *Exchange
	Route    config.Route
}

// registry filter for chain
var registryFilters = map[string]Filter{}

// RegistryFilter register filter
func RegistryFilter(filter Filter) {
	registryFilters[filter.Name()] = filter
}

// Init filter
func Init() {
	for _, filter := range registryFilters {
		filter.Init()
	}
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
		fc := c.Route.GetFilter(f.Name())
		err := f.Filter(c.Exchange, fc.Params)
		if err != nil {
			c.Exchange.Error(err)
			complete(c.Exchange)
		}

		if c.Exchange.Completed {
			complete(c.Exchange)
			return
		}
	}

}

func complete(exchange *Exchange) {
	if exchange.Err != nil {
		log.Error(exchange.Err)
	}
	log.Debug("complete")
}

func mapstruct(c interface{}, config interface{}) error {
	return mapstructure.Decode(c, config)
}
