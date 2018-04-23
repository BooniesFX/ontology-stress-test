package main

import (
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"syscall"

	"github.com/ontio/ontology-crypto/keypair"
	actor "github.com/ontio/ontology-stress-test/actor"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/http/jsonrpc"
	"github.com/ontio/ontology/http/restful"
	"github.com/ontio/ontology/p2pserver"
	netreqactor "github.com/ontio/ontology/p2pserver/actor/req"
)

const (
	DefaultMultiCoreNum = 4
)

func init() {
	log.Init(log.PATH, log.Stdout)

	var coreNum int
	if config.Parameters.MultiCoreNum > DefaultMultiCoreNum {
		coreNum = int(config.Parameters.MultiCoreNum)
	} else {
		coreNum = DefaultMultiCoreNum
	}
	log.Debug("The Core number is ", coreNum)
	runtime.GOMAXPROCS(coreNum)
}

func main() {
	var acct *account.Account
	var err error
	log.Trace("Node version: ", config.Version)

	consensusType := strings.ToLower(config.Parameters.ConsensusType)
	if consensusType == "dbft" && len(config.Parameters.Bookkeepers) < account.DEFAULT_BOOKKEEPER_COUNT {
		log.Fatal("With dbft consensus type, at least ", account.DEFAULT_BOOKKEEPER_COUNT, " Bookkeepers should be set in config.json")
		os.Exit(1)
	}

	log.Info("0. Open the account")
	client := account.GetClient()
	if client == nil {
		log.Fatal("Can't get local account.")
		os.Exit(1)
	}
	acct = client.GetDefaultAccount()
	if acct == nil {
		log.Fatal("can not get default account")
		os.Exit(1)
	}
	log.Debug("The Node's PublicKey ", acct.PublicKey)
	defBookkeepers, err := client.GetBookkeepers()
	sort.Sort(keypair.NewPublicList(defBookkeepers))
	if err != nil {
		log.Fatalf("GetBookkeepers error:%s", err)
		os.Exit(1)
	}
	//Init event hub
	events.Init()

	ldgerActor := actor.NewLedgerActor()
	ledgerPID := ldgerActor.Start()

	txnpoolActor := actor.NewTxnPoolActor()
	txnPID := txnpoolActor.Start()

	log.Info("Start the P2P networks")

	p2p, err := p2pserver.NewServer(acct)
	if err != nil {
		log.Fatalf("p2pserver NewServer error %s", err)
		os.Exit(1)
	}
	err = p2p.Start(false)
	if err != nil {
		log.Fatalf("p2p sevice start error %s", err)
		os.Exit(1)
	}
	netreqactor.SetLedgerPid(ledgerPID)
	netreqactor.SetTxnPoolPid(txnPID)
	log.Info("--Start the RPC interface")
	go jsonrpc.StartRPCServer()
	go restful.StartServer()

	p2p.WaitForPeersStart()
	log.Info("wait for test data...")
	go actor.LoopPrintActorInfo()
	//等待退出信号
	waitToExit()
}

func waitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("Ontology received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}
