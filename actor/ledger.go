package actor

import (
	"fmt"
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	ledger "github.com/ontio/ontology/core/ledger/actor"
)

var DefLedgerPid *actor.PID

type LedgerActor struct {
	props *actor.Props
}

func NewLedgerActor() *LedgerActor {
	return &LedgerActor{}
}

func (self *LedgerActor) Start() *actor.PID {
	self.props = actor.FromProducer(func() actor.Actor { return self })
	var err error
	DefLedgerPid, err = actor.SpawnNamed(self.props, "LedgerActor")
	if err != nil {
		panic(fmt.Errorf("LedgerActor SpawnNamed error:%s", err))
	}
	return DefLedgerPid
}

func (self *LedgerActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Stop:
	case *ledger.GetCurrentBlockHeightReq:
		self.handleGetCurrentBlockHeightReq(ctx, msg)
	case *ledger.GetCurrentHeaderHeightReq:
		self.handleGetCurrentHeaderHeightReq(ctx, msg)
	default:
		log.Warnf("LedgerActor cannot deal with type: %v %v", msg, reflect.TypeOf(msg))
	}
}

func (self *LedgerActor) handleGetCurrentBlockHeightReq(ctx actor.Context, req *ledger.GetCurrentBlockHeightReq) {
	curBlockHeight := uint32(1)
	resp := &ledger.GetCurrentBlockHeightRsp{
		Height: curBlockHeight,
		Error:  nil,
	}
	ctx.Sender().Request(resp, ctx.Sender())
}

func (self *LedgerActor) handleGetCurrentHeaderHeightReq(ctx actor.Context, req *ledger.GetCurrentHeaderHeightReq) {
	curHeaderHeight := uint32(1)
	resp := &ledger.GetCurrentHeaderHeightRsp{
		Height: curHeaderHeight,
		Error:  nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}
