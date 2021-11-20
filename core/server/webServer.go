package server

import (
	"context"
	"errors"
	"github.com/cyejing/shuttle/core/filter"
	"github.com/cyejing/shuttle/pkg/common"
	"github.com/cyejing/shuttle/pkg/config"
	"github.com/cyejing/shuttle/pkg/log"
	"net"
	"net/http"
	"sort"
	"strings"
)

type RouteMux struct {
	Routes []config.Route
}

func StartWebServer() {
	c := config.GetConfig()

	routeMux := &RouteMux{Routes: c.Routes}

	sort.Slice(routeMux.Routes, func(i, j int) bool {
		return routeMux.Routes[i].Order > routeMux.Routes[j].Order
	})

	ser := &http.Server{
		Addr:    c.Addr,
		Handler: routeMux,
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(ctx, common.ConnContextKey, c)
		},
	}

	filter.Init()

	ctx := context.Background()
	errChan := make(chan error)

	if c.Ssl.Enable {
		go func() {
			err := ser.ListenAndServeTLS(c.Ssl.Cert, c.Ssl.Key)
			if err != nil {
				log.Error("启动服务失败,请检查证书配置文件", err)
				errChan <- err
			}
			log.Infof("Start TLS Web Server for addr %s", c.Ssl.Addr)
		}()
	}

	go func() {
		err := ser.ListenAndServe()
		if err != nil {
			log.Error("启动服务失败", err)
			errChan <- err
		}
		log.Infof("Start Web Server for addr %s", c.Addr)
	}()

	select {
	case err := <-errChan:
		log.Panic(err)
	case <-ctx.Done():
		log.Info("ctx done exit")
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
