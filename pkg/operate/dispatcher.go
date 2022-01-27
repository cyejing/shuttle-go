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

var typeMap = make(map[Type]func() Operate)

func registerOp(t Type, newOP func() Operate) {
	typeMap[t] = newOP
}

// Dispatcher struct
type Dispatcher struct {
	ReqMap   *sync.Map
	Wormhole *Wormhole
	Channel  chan Operate
}

func NewDispatcher(wormhole *Wormhole) *Dispatcher {
	return &Dispatcher{
		ReqMap:   &sync.Map{},
		Wormhole: wormhole,
		Channel:  make(chan Operate, 10),
	}
}

func (d *Dispatcher) Run() error {
	go func() {
		err := d.Dispatch()
		if err != nil {
			log.Error(err)
		}
	}()
	return d.Read()
}

func (d *Dispatcher) Exchange() error {
	d.Send(NewExchangeOP(d.Wormhole.Name, nil))
	return d.Run()
}

func (d *Dispatcher) Send(c Operate) {
	if !c.IsResponse() {
		d.ReqMap.Store(c.GetReqId(), c)
	}
	d.Channel <- c
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

func extractDispatcher(ctx context.Context) (*Dispatcher, error) {
	if d, ok := ctx.Value(common.DispatcherKey).(*Dispatcher); ok {
		return d, nil
	} else {
		return nil, utils.NewErrf("cannot expect context:%v", ctx.Value(common.DispatcherKey))
	}
}
