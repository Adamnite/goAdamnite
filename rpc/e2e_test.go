package rpc

import (
	"fmt"
	"log"
	"math/big"
	"net/rpc"
	"os"
	"testing"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
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
	testDB        = rawdb.NewMemoryDB()
	stateDB, _    = statedb.New(common.Hash{}, statedb.NewDatabase(testDB))
	chainConfig   = params.TestnetChainConfig
	client        AdamniteClient
	bouncerServer *BouncerServer
	bouncerClient *rpc.Client
)

func setup() {
	// setup Adamnite server
	for i, address := range testBalances {
		testBalances[i] = big.NewInt(0).Mul(niteBigExponent, address)
		stateDB.AddBalance(testAccounts[i], testBalances[i])
	}

	rootHash := stateDB.IntermediateRoot(false)
	stateDB.Database().TrieDB().Commit(rootHash, false, nil)

	blockchain, err := blockchain.NewBlockchain(
		testDB,
		chainConfig,
		dpos.New(chainConfig, testDB),
	)

	if err != nil {
		log.Printf("[Adamnite E2E test] Error: %s", err)
		return
	}

	var port uint32 = 12345
	var bouncerPort uint32 = 12346

	bouncerServer = NewBouncerServer(stateDB, blockchain, bouncerPort)
	adamniteServer := NewAdamniteServer(port)
	go adamniteServer.Run()

	defer func() {
		adamniteServer.Close()
		bouncerServer.Close()
	}()

	// setup Adamnite client
	client, err = NewAdamniteClient(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Printf("[Adamnite E2E test] Error: %s", err)
		return
	}
	bouncerClient, err = rpc.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", bouncerPort))
	if err != nil {
		log.Printf("[Adamnite E2E test] Bouncer Error: %s", err)
		return
	}
}

func shutdown() {
	client.Close()
	bouncerClient.Close()
}

func TestGetVersion(t *testing.T) {
	leeway := time.Second / 10 //no actions can be instant, so this is how much time allowance i give.
	client.SetAddressAndHostingPort(&common.Address{123}, "")
	version, err := client.GetVersion()
	now := time.Now().UTC()
	if err != nil {
		t.Fatal(err)
	}
	//timestamp is going to be off, but shouldn't be too off
	assert.Equal(t, now.Round(leeway), version.Timestamp.Round(leeway), "time is too far off")
	//TODO: check the rest of this is indeed working
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}
