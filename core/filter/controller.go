package filter

import (
	controller2 "github.com/cyejing/shuttle/core/controller"
	"github.com/cyejing/shuttle/pkg/config/server"
)

type controller struct {
	name string
}

func init() {
	RegistryFilter(&controller{name: "controller"})
}

func (s controller) Init(mux *RouteMux) {
	mux.Routes = append(mux.Routes, server.Route{
		ID:    "controller-filter",
		Order: 1000,
		Path:  "/ctl/.*",
		Filters: []server.Filter{
			{
				Name:     "controller",
				Params:   nil,
				Open:     true,
				Loggable: false,
			},
		},
		Loggable: false,
	})

}

func (s controller) Name() string {
	return s.name
}

func (s controller) Filter(exchange *Exchange, c interface{}) error {
	path := exchange.Req.URL.Path
	switch path {
	case "/ctl/proxy":
		wormholeName := exchange.Req.FormValue("wormholeName")
		shipName := exchange.Req.FormValue("shipName")
		remoteAddr := exchange.Req.FormValue("remoteAddr")
		localAddr := exchange.Req.FormValue("localAddr")
		if wormholeName == "" || shipName == "" || remoteAddr == "" || localAddr == "" {
			println("error params")
		}else{
			println(wormholeName,shipName,remoteAddr,localAddr)
			go func() {
				err := controller2.NewProxyCtl(wormholeName, shipName, remoteAddr, localAddr).Run()
				if err != nil {
					log.Errorln(err)
				}
			}()
		}
	default:
		log.Infof("controller no impl path %s", path)
	}
	return nil
}
