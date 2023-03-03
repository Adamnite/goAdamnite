package rpc

import (
	"fmt"
	"math/big"
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
	testAddress = common.BytesToAddress([]byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11, 0x12, 0x13})
	testBalance = big.NewInt(1).Mul(big.NewInt(9000000000000000000), big.NewInt(1000))
	test_db     = rawdb.NewMemoryDB()
	state, _    = statedb.New(common.Hash{}, statedb.NewDatabase(test_db))
	chainConfig = params.TestnetChainConfig
)

func setupTestingServer() AdamniteServer {
	state.AddBalance(testAddress, testBalance)
	rootHash := state.IntermediateRoot(false)
	state.Database().TrieDB().Commit(rootHash, false, nil)

	bc, err := core.NewBlockchain(test_db,
		chainConfig,
		dpos.New(chainConfig, test_db),
	)
	if err != nil {
		fmt.Println(err)
	}

	admServer := NewAdamniteServer(state, bc)
	admServer.Launch()
	return *admServer
}
func TestGetBalance(t *testing.T) {
	// t.Parallel()
	admServer := setupTestingServer()

	// fmt.Println(admServer.Endpoint)
	client := NewAdamniteClient(admServer.Endpoint)

	value, err := client.GetBalance(testAddress)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	if !assert.Equal(t, testBalance, value, "balances are not matching after RPC call.") {
		t.Fail()
	}

}
func TestGetChainID(t *testing.T) {
	// t.Parallel()
	admServer := setupTestingServer()

	client := NewAdamniteClient(admServer.Endpoint)

	value, err := client.GetChainID()
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	if !assert.Equal(t, chainConfig.ChainID, value, "chain ids are not matching after RPC call.") {
		t.Fail()
	}
}
