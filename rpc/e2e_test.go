package rpc

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/params"
	"github.com/stretchr/testify/assert"
)

var (
	niteBigExponent = big.NewInt(1).Exp(big.NewInt(10), big.NewInt(20), nil) // big math version of 10**20
	testAccounts    = []common.Address{
		common.BytesToAddress([]byte{0x00}),
		common.BytesToAddress([]byte{0x01}),
		common.BytesToAddress([]byte{0x02}),
	}
	testBalances = []*big.Int{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(2),
	}
	testDB      = rawdb.NewMemoryDB()
	stateDB, _  = statedb.New(common.Hash{}, statedb.NewDatabase(testDB))
	chainConfig = params.TestnetChainConfig
	client      AdamniteClient
)

func setup() {
	// setup Adamnite server
	for i, address := range testBalances {
		testBalances[i] = big.NewInt(0).Mul(niteBigExponent, address)
		stateDB.AddBalance(testAccounts[i], testBalances[i])
	}

	rootHash := stateDB.IntermediateRoot(false)
	stateDB.Database().TrieDB().Commit(rootHash, false, nil)

	blockchain, err := core.NewBlockchain(
		testDB,
		chainConfig,
		dpos.New(chainConfig, testDB),
	)

	if err != nil {
		log.Printf("[Adamnite E2E test] Error: %s", err)
		return
	}

	var port uint32
	port = 12345

	adamniteServer := NewAdamniteServer(stateDB, blockchain, port)
	defer func() {
		adamniteServer.Close()
	}()
	go adamniteServer.Run()

	// setup Adamnite client
	client, err = NewAdamniteClient(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Printf("[Adamnite E2E test] Error: %s", err)
		return
	}
}

func shutdown() {
	client.Close()
}

func TestGetBalance(t *testing.T) {
	balance, err := client.GetBalance(testAccounts[0])
	if err != nil {
		log.Printf("[Adamnite E2E test] Error: %s", err)
		t.Fail()
	}
	if !assert.Equal(t, testBalances[0].String(), *balance, "Balances do not match") {
		t.Fail()
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}
