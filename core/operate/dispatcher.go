package operate

import (
	"bytes"
	"context"
	"github.com/cyejing/shuttle/pkg/common"
	"github.com/cyejing/shuttle/pkg/errors"
	"github.com/cyejing/shuttle/pkg/logger"
	"reflect"
	"sync"
)

var log = logger.NewLog()

var DispatcherMap = &sync.Map{}

func GetCliDispatcher(name string) *Dispatcher {
	return getDispatcher("Cli-" + name)
}

func GetSerDispatcher(name string) *Dispatcher {
	return getDispatcher("Ser-" + name)
}

func getDispatcher(name string) *Dispatcher {
	if a, ok := DispatcherMap.Load(name); ok {
		if d, o := a.(*Dispatcher); o {
			return d
		}
	}
	return nil
}

var typeMap = make(map[Type]func() Operate)

func registerOp(t Type, newOP func() Operate) {
	typeMap[t] = newOP
}

// Dispatcher struct
type Dispatcher struct {
	Name        string
	Key         string
	reqMap      *sync.Map
	ExchangeMap *sync.Map
	ProxyMap    *sync.Map
	Wormhole    *Wormhole
	Channel     chan Operate
}

func NewCliDispatcher(wormhole *Wormhole, name string) *Dispatcher {
	return newDispatcher(wormhole, name, "Cli-"+name)
}

func NewSerDispatcher(wormhole *Wormhole, name string) *Dispatcher {
	return newDispatcher(wormhole, name, "Ser-"+name)
}

func newDispatcher(wormhole *Wormhole, name string, key string) *Dispatcher {
	d := &Dispatcher{
		Name:        name,
		Key:         key,
		reqMap:      &sync.Map{},
		ExchangeMap: &sync.Map{},
		ProxyMap:    &sync.Map{},
		Wormhole:    wormhole,
		Channel:     make(chan Operate, 10),
	}
	DispatcherMap.Store(key, d)
	return d
}

func (d *Dispatcher) Run() error {
	errChan := make(chan error, 2)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		err := d.Dispatch(ctx)
		errChan <- err
	}()
	go func() {
		err := d.Read()
		errChan <- err
	}()
	err := <-errChan
	cancelFunc()
	d.Wormhole.Rwc.Close()
	d.clean()
	return err
}

func (d *Dispatcher) Connect() error {
	connectOP := NewConnectOP(d.Name)
	connectOP.respCall = func(req *ReqBase, resp *RespOP) {
		err := scanProxyConfig(d)
		if err != nil {
			log.Error(err)
		}
	}
	d.Send(connectOP)
	return d.Run()
}

func (d *Dispatcher) Send(o Operate) {
	if req, ok := o.(ReqOperate); ok {
		d.reqMap.Store(req.GetReqBase().reqId, req)
	}
	d.Channel <- o
}
func (d *Dispatcher) SendAndWait(req ReqOperate) *RespOP {
	d.reqMap.Store(req.GetReqBase().reqId, req)
	d.Channel <- req
	return req.WaitResp()
}

func (d *Dispatcher) Dispatch(ctx context.Context) error {
	buf := bytes.NewBuffer([]byte{})
	for {
		buf.Reset()
		select {
		case c := <-d.Channel:
			err := c.Encode(buf)
			if err != nil {
				log.Error(errors.BaseErr("command encode err", err))
			}
			_, err = d.Wormhole.Rwc.Write(buf.Bytes())
			if err != nil {
				return errors.BaseErr("handle ReqBase write byte fail", err)
			}
		case <-ctx.Done():
			return errors.NewErrf("dispatch done %s", d.Name)
		}
	}
}

func (d *Dispatcher) Read() error {
	for {
		buf := d.Wormhole.Br
		tb, err := buf.Peek(1)
		if err != nil {
			return err
		}
		t := Type(tb[0])

		newOp := typeMap[t]
		if newOp == nil {
			buf.Discard(1)
			log.Error(errors.NewErrf("unknown type: %v", tb[0]))
			continue
		}
		op := newOp()
		//log.Infof("%s read op %s, remote[%v]", d.Key, reflect.TypeOf(op).String(), d.Wormhole.Rwc.RemoteAddr())

		err = op.Decode(buf)
		if err != nil {
			log.Error(errors.BaseErrf("op %s decode err", err, reflect.TypeOf(op).Name()))
			continue
		}
		err = op.Execute(context.WithValue(context.Background(), common.DispatcherKey, d))
		if err != nil {
			log.Error(errors.BaseErr("command execute err", err))
		}
	}
}

func (d *Dispatcher) LoadReq(reqId uint32) (ReqOperate, bool) {
	if a, ok := d.reqMap.LoadAndDelete(reqId); ok {
		if req, o := a.(ReqOperate); o {
			return req, true
		}
	}
	return nil, false
}

func (d *Dispatcher) LoadExchange(name string) (ExchangeCtl, bool) {
	if a, ok := d.ExchangeMap.Load(name); ok {
		if ec, o := a.(ExchangeCtl); o {
			return ec, true
		}
	}
	return nil, false
}

func (d *Dispatcher) DeleteExchange(name string) {
	d.ExchangeMap.Delete(name)
}

func (d *Dispatcher) clean() {
	log.Infof("clean dispatcher :%s", d.Key)
	DispatcherMap.Delete(d.Key)
	d.ProxyMap.Range(func(key, value interface{}) bool {
		if p, ok := value.(*ProxyCtl); ok {
			p.Stop()
		}
		return true
	})
}

func extractDispatcher(ctx context.Context) (*Dispatcher, error) {
	if d, ok := ctx.Value(common.DispatcherKey).(*Dispatcher); ok {
		return d, nil
	} else {
		return nil, errors.NewErrf("cannot expect context:%v", ctx.Value(common.DispatcherKey))
	}
}
