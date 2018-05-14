package actor

import (
	"fmt"
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
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
	default:
		log.Warnf("LedgerActor cannot deal with type: %v %v", msg, reflect.TypeOf(msg))
	}
}
