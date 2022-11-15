package main

import (
	"context"
	"crypto"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/adamnite/go-adamnite/accounts"
	"github.com/adamnite/go-adamnite/accounts/keystore"
	"github.com/adamnite/go-adamnite/cmd/utils"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	var (
		senderAddr      = flag.String("sendaddr", "", "sender address")
		receiptAddr     = flag.String("recaddr", "", "receipt address")
		transactionType = flag.Bool("txtype", false, "default is false, true: voting transaction, false: transaction")
		niteAmount      = flag.Int("amount", 0, "amount to send")
		accountBalance  = flag.String("balance", "", "balance of address")
		keyfile         = flag.String("keyfile", "", "key file of sender address")
		password        = flag.String("password", "", "password of key file")
	)
	flag.Parse()

	if *accountBalance == "" {
		if *senderAddr == "" || *receiptAddr == "" || *niteAmount == 0 || *keyfile == "" {
			switch {
			case *senderAddr == "":
				utils.Fatalf("sender address should be inputed")
			case *receiptAddr == "":
				utils.Fatalf("receipt address should be inputed.")
			case *niteAmount == 0:
				utils.Fatalf("Nite coin amount should be greater than zero.")
			case *keyfile != "":
				utils.Fatalf("keystore file should be existed.")
			}
			os.Exit(0)
		}

		acctAddr, _ := crypto.B58decode(*senderAddr)
		sendAddr := accounts.Account{Address: common.BytesToAddress(acctAddr)}

		accntAddr, _ := crypto.B58decode(*receiptAddr)
		recAddr := accounts.Account{Address: common.BytesToAddress(accntAddr)}

		ks := keystore.NewKeyStore("./tmp", keystore.StandardScryptN, keystore.StandardScryptP)
		jsonBytes, err := ioutil.ReadFile(*keyfile)
		if err != nil {
			utils.Fatalf("keystore file should be existed.")
		}

		account, err := ks.Import(jsonBytes, *password, *password)
		if err != nil {
			utils.Fatalf(err.Error())
		}

		makeTransaction(sendAddr.Address, recAddr.Address, *niteAmount, *transactionType, account, ks, *password)
	} else {
		acctAddr, _ := crypto.B58decode(*accountBalance)

		client, err := ethclient.Dial("\\\\.\\pipe\\nite.ipc")
		if err != nil {
			utils.Fatalf("Network failed")
			utils.Fatalf(err.Error())
		}

		balance, err := client.BalanceAt(context.Background(), common.BytesToAddress(acctAddr), nil)
		if err != nil {
			utils.Fatalf("Get Balance Error")
		}
		utils.Infof(balance.String())

	}

}

func makeTransaction(from common.Address, to common.Address, amount int, txtype bool, accnt accounts.Account, ks *keystore.KeyStore, password string) {

	client, err := adamniteClient.New("\\\\.\\pipe\\nite.ipc")

	if err != nil {
		utils.Fatalf("Network failed")
		utils.Fatalf(err.Error())
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		utils.Fatalf(err.Error())
	}
	nonce, _ := client.PendingNonceAt(context.Background(), from)

	tx := types.Transaction(nonce, to, big.NewInt(int64(amount*1000000000000000000)), uint64(21000), big.NewInt(30000000000), nil)
	if txtype {
		tx = types.Transaction(nonce, crypto.BLAKE2b_256.HashFunc(), big.NewInt(0), uint64(750000), gasPrice, []byte("vote:5000000"))
	}
	chainId, _ := client.ChainID(context.Background())
	signedTx, _ := ks.SignTxWithPassphrase(accnt, password, tx, chainId)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		utils.Fatalf(err.Error())
	}
	fmt.Printf("tx sent: %s", tx.Hash().Hex())

}
