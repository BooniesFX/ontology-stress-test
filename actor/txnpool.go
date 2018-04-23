package actor

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	tc "github.com/ontio/ontology/txnpool/common"
)

var TxCnt uint64
var TxCntLatest uint64

var DefTxnPid *actor.PID

type TxnPoolActor struct {
	props *actor.Props
}

func NewTxnPoolActor() *TxnPoolActor {
	return &TxnPoolActor{}
}

func (self *TxnPoolActor) Start() *actor.PID {
	self.props = actor.FromProducer(func() actor.Actor { return self })
	var err error
	DefTxnPid, err = actor.SpawnNamed(self.props, "TxnPoolActor")
	if err != nil {
		panic(fmt.Errorf("TxnPoolActor SpawnNamed error:%s", err))
	}
	return DefTxnPid
}

func (self *TxnPoolActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Stop:
	case *tc.TxReq:
		AddTransaction(msg.Tx)
	case *tc.GetTxnReq:
		sender := ctx.Sender()
		if sender != nil {
			sender.Request(&tc.GetTxnRsp{Txn: nil},
				ctx.Self())
		}
	default:
		log.Warnf("TxnPoolActor cannot deal with type: %v %v", msg, reflect.TypeOf(msg))
	}
}

func AddTransaction(transaction *types.Transaction) {
	atomic.AddUint64(&(TxCnt), 1)
}

func PrintTxnInfo() {
	txnPerSnd := TxCnt - TxCntLatest
	TxCntLatest = TxCnt
	fmt.Printf("total txn count %d,TPS = %d/s\n", TxCnt, txnPerSnd)
}

func LoopPrintActorInfo() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			PrintTxnInfo()
		}
	}
}
