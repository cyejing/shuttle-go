package operate

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/cyejing/shuttle/pkg/common"
	"github.com/cyejing/shuttle/pkg/logger"
	"github.com/cyejing/shuttle/pkg/utils"
	"reflect"
	"sync"
)

var log = logger.NewLog()

var dispatcherMap = &sync.Map{}

func GetCliDispatcher(name string) *Dispatcher {
	return getDispatcher("Cli-" + name)
}

func GetSerDispatcher(name string) *Dispatcher {
	return getDispatcher("Ser-" + name)
}

func getDispatcher(name string) *Dispatcher {
	if a, ok := dispatcherMap.Load(name); ok {
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
	exchangeMap *sync.Map
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
		exchangeMap: &sync.Map{},
		Wormhole:    wormhole,
		Channel:     make(chan Operate, 10),
	}
	dispatcherMap.Store(key, d)
	return d
}

func (d *Dispatcher) Run() error {
	errChan := make(chan error)
	go func() {
		err := d.Dispatch()
		errChan <- err
	}()
	go func() {
		err := d.Read()
		errChan <- err
	}()
	err := <-errChan
	if err != nil {
		log.Error(err)
	}
	dispatcherMap.Delete(d.Key)
	return err
}

func (d *Dispatcher) Connect() error {
	d.Send(NewConnectOP(d.Name))
	return d.Run()
}

func (d *Dispatcher) Send(o Operate) {
	if req, ok := o.(ReqOperate); ok {
		d.reqMap.Store(req.GetReqBase().reqId, req)
	}
	d.Channel <- o
}

func (d *Dispatcher) Dispatch() error {
	buf := bytes.NewBuffer([]byte{})
	for {
		buf.Reset()
		select {
		case c := <-d.Channel:
			err := c.Encode(buf)
			if err != nil {
				log.Error(utils.BaseErr("command encode err", err))
			}
			fmt.Println(hex.Dump(buf.Bytes()))
			_, err = d.Wormhole.Rwc.Write(buf.Bytes())
			if err != nil {
				return utils.BaseErr("handle ReqBase write byte fail", err)
			}
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
			log.Error(utils.NewErrf("unknow type op: %v", t))
		}
		op := newOp()

		err = op.Decode(buf)
		if err != nil {
			log.Error(utils.BaseErrf("op %s decode err", err, reflect.TypeOf(op).Name()))
		}
		err = op.Execute(context.WithValue(context.Background(), common.DispatcherKey, d))
		if err != nil {
			log.Error(utils.BaseErr("command execute err", err))
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

func (d *Dispatcher) LoadExchange(name string) (ExchangeConn, bool) {
	if a, ok := d.exchangeMap.Load(name); ok {
		if ec, o := a.(ExchangeConn); o {
			return ec, true
		}
	}
	return nil, false
}

func (d *Dispatcher) DeleteExchange(name string) {
	d.exchangeMap.Delete(name)
}

func extractDispatcher(ctx context.Context) (*Dispatcher, error) {
	if d, ok := ctx.Value(common.DispatcherKey).(*Dispatcher); ok {
		return d, nil
	} else {
		return nil, utils.NewErrf("cannot expect context:%v", ctx.Value(common.DispatcherKey))
	}
}
