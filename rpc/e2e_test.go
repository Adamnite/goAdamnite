package rpc

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/stretchr/testify/assert"
)

var (
	testAddress = common.BytesToAddress([]byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11, 0x12, 0x13})
)

func TestServerDirectly(t *testing.T) {
	testBalance := big.NewInt(9000000000000000000)
	testBalance.Mul(testBalance, big.NewInt(100))

	db := rawdb.NewMemoryDB()
	state, _ := statedb.New(common.Hash{}, statedb.NewDatabase(db))
	state.AddBalance(testAddress, testBalance)
	rootHash := state.IntermediateRoot(false)
	state.Database().TrieDB().Commit(rootHash, false, nil)

	admServer := NewAdamniteServer(*state)
	admServer.Launch()

	returnedInt := BigIntReply{}
	admServer.GetBalance(testAddress, &returnedInt)
	assert.Equal(t, testBalance, returnedInt.toBigInt())
}
func TestGetBalance(t *testing.T) {
	testBalance := big.NewInt(9000000000000000000)
	testBalance.Mul(testBalance, big.NewInt(100))

	db := rawdb.NewMemoryDB()
	state, _ := statedb.New(common.Hash{}, statedb.NewDatabase(db))
	state.AddBalance(testAddress, testBalance)
	rootHash := state.IntermediateRoot(false)
	state.Database().TrieDB().Commit(rootHash, false, nil)

	admServer := NewAdamniteServer(*state)
	admServer.Launch()
	fmt.Println(admServer.Endpoint)
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
