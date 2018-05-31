package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

var (
	COUNT       int
	TPS         int
	WORKER      int
	RPC         string
	TO          string
	WALLET_FILE string
	WALLET_PWD  string
)

var (
	OntSdk *sdk.OntologySdk
	Admin  *account.Account
)

func init() {
	flag.IntVar(&COUNT, "r", 100000, "Request count")
	flag.IntVar(&TPS, "tps", 1000, "tx per second")
	flag.IntVar(&WORKER, "w", 10, "Worker num")
	flag.StringVar(&RPC, "rpc", "http://localhost:20336", "Default address of ontology rpc")
	flag.StringVar(&TO, "to", "", "Dest address")
	flag.StringVar(&WALLET_FILE, "wallet", "./wallet.dat", "Wallet file path")
	flag.StringVar(&WALLET_PWD, "pwd", "pwd", "Password of wallet")
	flag.Parse()
}

func main() {
	log.InitLog(log.InfoLog)
	OntSdk = sdk.NewOntologySdk()
	OntSdk.Rpc.SetAddress(RPC)
	wallet, err := OntSdk.OpenWallet(WALLET_FILE)
	if err != nil {
		fmt.Printf("OpenWallet error:%s\n", err)
		return
	}
	Admin, err = wallet.GetDefaultAccount([]byte(WALLET_PWD))
	if err != nil {
		fmt.Printf("CreateAccount error:%s", err)
		return
	}
	fmt.Printf("Admin:%x\n", keypair.SerializePublicKey(Admin.PublicKey))

	balance, err := OntSdk.Rpc.GetBalance(Admin.Address)
	if err != nil {
		fmt.Printf("GetBalance error:%s\n", err)
		return
	}

	fmt.Printf("Admin ont balance:%d\n", balance.Ont)
	if balance.Ont < 0 {
		fmt.Printf("Admin balance not enought\n")
		return
	}

	if TO == "" {
		fmt.Println("Dest address should not be nil")
		return
	}
	TestTransfer()
	balance, err = OntSdk.Rpc.GetBalance(Admin.Address)
	if err != nil {
		fmt.Printf("GetBalance error:%s\n", err)
		return
	}
	<-time.After(time.Second * 3)
	fmt.Printf("Admin ont left:%d\n", balance.Ont)
	if balance.Ont < 0 {
		fmt.Printf("Admin balance not enought\n")
		return
	}
}
func TestTransfer() {
	taskCh := make(chan int, WORKER)
	timerCh := make(chan int, 1)
	toAcc, _ := common.AddressFromBase58(TO)
	workerid := 0
	index := 0
	work := func() {
		for {
			select {
			case t := <-taskCh:
				if t == 0 {
					workerid++
					fmt.Printf("worker %d done:%v\n", workerid, time.Now())
					if workerid == WORKER {
						timerCh <- 0
					}
					return
				}
				index++
				_, err := OntSdk.Rpc.Transfer(0, 30000+uint64(index), "ont", Admin, toAcc, 1)
				if err != nil {
					fmt.Printf("transfer error:%s\n", err)
					return
				}
			}
		}
	}

	for i := 0; i < WORKER; i++ {
		go work()
	}

	reqCount := 0
	timer := time.NewTicker(time.Second)
	for {
		select {
		case <-timer.C:
			fmt.Printf("Transfer start:%v\n", time.Now())
			for i := 0; i < TPS; i++ {
				taskCh <- 1
				reqCount++
				if reqCount == COUNT {
					for i := 0; i < WORKER; i++ {
						taskCh <- 0
						timer.Stop()
					}
					break
				}
			}
			fmt.Printf("transfer complete:%d\n", reqCount)
		case t := <-timerCh:
			if t == 0 {
				return
			}
		}

	}
}
