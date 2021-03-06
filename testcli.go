package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"sort"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	//ldgactor "github.com/ontio/ontology-stress-test/actor"
	"github.com/ontio/ontology/account"
	_ "github.com/ontio/ontology/cli"
	test "github.com/ontio/ontology/cli/test"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
)

var Version string

var (
	Ip   string
	Port string
)

func main() {
	log.Init(log.Stdout)
	app := cli.NewApp()
	app.Name = "nodectl"
	app.Version = Version
	app.HelpName = "testcli"
	app.Usage = "command line tool for Ontology stress test"
	app.UsageText = "nodectl [global options] command [command options] [args]"
	app.HideHelp = false
	app.HideVersion = false
	//global options
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "ip",
			Usage:       "node's ip address",
			Value:       "127.0.0.1",
			Destination: &Ip,
		},
		cli.StringFlag{
			Name:        "port",
			Usage:       "node's port",
			Value:       "20338",
			Destination: &Port,
		},
	}
	//commands
	app.Commands = []cli.Command{
		*NewCommand(),
	}
	sort.Sort(cli.CommandsByName(app.Commands))
	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)
}

func signTransaction(signer *account.Account, tx *types.Transaction) error {
	hash := tx.Hash()
	sign, _ := signature.Sign(signer, hash[:])
	tx.Sigs = append(tx.Sigs, &types.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sign},
	})
	return nil
}

func testAction(c *cli.Context) (err error) {
	txnNum := c.Int("num")
	passwd := c.String("password")
	genFile := c.Bool("gen")
	acct := account.Open("wallet.dat", []byte(passwd))
	if acct == nil {
		fmt.Println(" can not get default account")
		os.Exit(1)
	}
	acc := acct.GetDefaultAccount()
	if acc == nil {
		fmt.Println(" can not get default account")
		os.Exit(1)
	}
	if genFile {
		GenTransferFile(txnNum, acc, "transfer.dat")
		return nil
	}
	fmt.Println("start to connect destination peer...")
	//connect
	//fake ledger
	//ldgerActor := ldgactor.NewLedgerActor()
	//ledgerPID := ldgerActor.Start()

	racc := account.NewAccount("SHA256withECDSA")
	p, _ := p2pserver.NewServer(racc)
	p.Start()
	defer p.Stop()
	//netreqactor.SetLedgerPid(ledgerPID)

	nodeAddr := Ip + ":" + Port
	p.GetNetWork().Connect(nodeAddr, false)
	<-time.After(time.Second * 5)
	if p.GetConnectionCnt() >= 1 {

		fmt.Println("peer connected, begin test process")
		transferTest(txnNum, acc, p)
	}
	return nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func Tx2Hex(tx *types.Transaction) string {
	var buffer bytes.Buffer
	tx.Serialize(&buffer)
	return hex.EncodeToString(buffer.Bytes())
}

func GenTransferFile(n int, acc *account.Account, fileName string) {
	f, err := os.Create(fileName)
	check(err)
	w := bufio.NewWriter(f)

	defer func() {
		w.Flush()
		f.Close()
	}()

	for i := 0; i < n; i++ {
		to := acc.Address
		binary.BigEndian.PutUint64(to[:], uint64(i))
		tx := test.NewOntTransferTransaction(acc.Address, to, 1)
		if err := signTransaction(acc, tx); err != nil {
			fmt.Println("signTransaction error:", err)
			os.Exit(1)
		}

		txhex := Tx2Hex(tx)
		_, _ = w.WriteString(fmt.Sprintf("%x,%s\n", tx.Hash(), txhex))
	}

}

func transferTest(n int, acc *account.Account, p *p2pserver.P2PServer) {
	if n <= 0 {
		n = 1
	}

	txn := test.NewOntTransferTransaction(acc.Address, acc.Address, 1)
	if err := signTransaction(acc, txn); err != nil {
		fmt.Println("signTransaction error:", err)
		os.Exit(1)
	}

	buffer, err := msgpack.NewTxn(txn)
	if err != nil {
		fmt.Println("Error New Tx message: ", err)
	}

	server := p.GetNetWork()
	fmt.Printf("%v - send test transation start\n", time.Now())
	for i := 0; i < n; i++ {
		server.Xmit(buffer, false)
	}
	fmt.Printf("%v - %d test transations done\n", time.Now(), n)
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "test",
		Usage:       "run test routine",
		Description: "With nodectl test, you could run simple tests.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "num, n",
				Usage: "sample transaction numbers",
				Value: 1,
			},
			cli.StringFlag{
				Name:  "password, p",
				Usage: "wallet password",
				Value: "passwordtest",
			},
			cli.BoolFlag{
				Name:  "gen, g",
				Usage: "gen transaction to file",
			},
		},
		Action: testAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			return cli.NewExitError("", 1)
		},
	}
}
