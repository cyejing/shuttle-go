package filter

import (
	"encoding/json"
	"fmt"
	"github.com/cyejing/shuttle/core/config/server"
	operate2 "github.com/cyejing/shuttle/core/operate"
	"github.com/cyejing/shuttle/pkg/utils"
	"net/http"
)

type controller struct {
	name string
}

func init() {
	RegistryFilter(&controller{name: "controller"})
}

func (s controller) Name() string {
	return s.name
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

	handler("/ctl/proxy.open", openProxy)
	handler("/ctl/wormhole.list", listWormhole)
}

type wormholeResult struct {
	Name     string
	Ship     []*shipResult
	Exchange []string
}
type shipResult struct {
	ShipName   string
	RemoteAddr string
	LocalAddr  string
}

func listWormhole(req *http.Request, r map[string]interface{}) {
	wormholes := make([]*wormholeResult, 0)
	operate2.DispatcherMap.Range(func(key, value interface{}) bool {
		if d, ok := value.(*operate2.Dispatcher); ok {
			ship := make([]*shipResult, 0)
			d.ProxyMap.Range(func(key, value interface{}) bool {
				if s, ok := value.(*operate2.ProxyCtl); ok {
					ship = append(ship, &shipResult{ShipName: s.ShipName, RemoteAddr: s.RemoteAddr, LocalAddr: s.LocalAddr})
				}
				return true
			})
			ex := make([]string, 0)
			d.ExchangeMap.Range(func(key, value interface{}) bool {
				if e, ok := value.(*operate2.ExchangeCtlStu); ok {
					ex = append(ex, fmt.Sprintf("exchange [%s] : %v -> %v", e.Name, e.Raw.LocalAddr(), e.Raw.RemoteAddr()))
				}
				return true
			})
			wormholes = append(wormholes, &wormholeResult{Ship: ship, Name: d.Name, Exchange: ex})
		}
		return true
	})
	r["wormholes"] = wormholes
}

func openProxy(req *http.Request, r map[string]interface{}) {
	wormholeName := req.FormValue("wormholeName")
	shipName := req.FormValue("ShipName")
	remoteAddr := req.FormValue("remoteAddr")
	localAddr := req.FormValue("localAddr")
	var msg string
	if wormholeName == "" || shipName == "" || remoteAddr == "" || localAddr == "" {
		msg = "error params"
	} else {
		dispatcher := operate2.GetSerDispatcher(wormholeName)
		if dispatcher == nil {
			msg = fmt.Sprintf("wormholeName %s does not exist", wormholeName)
			return
		}
		operate2.NewProxyCtl(dispatcher, shipName, remoteAddr, localAddr).Run()
		msg = "ok"
	}
	r["msg"] = msg
}

var pathMapping = make(map[string]func(req *http.Request, r map[string]interface{}))

func handler(path string, h func(req *http.Request, r map[string]interface{})) {
	pathMapping[path] = h
}
func (s controller) Filter(exchange *Exchange, c interface{}) error {
	path := exchange.Req.URL.Path
	h := pathMapping[path]
	var result = make(map[string]interface{})
	if h != nil {
		h(exchange.Req, result)
	}
	err := writeResult(exchange.Resp, result)
	if err != nil {
		return err
	}
	exchange.Completed()

	return nil
}

func writeResult(resp http.ResponseWriter, result map[string]interface{}) error {
	body := map[string]interface{}{"code": 0, "result": result}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return utils.BaseErr("json marshal err", err)
	}
	_, err = resp.Write(bodyBytes)
	return err
}
