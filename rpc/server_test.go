package rpc

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/stretchr/testify/assert"
)

func TestServerGetBalance(t *testing.T) {
	testBalance := big.NewInt(9000000000000000000)
	testBalance.Mul(testBalance, big.NewInt(100))

	db := rawdb.NewMemoryDB()
	state, _ := statedb.New(common.Hash{}, statedb.NewDatabase(db))
	state.AddBalance(testAddress, testBalance)
	rootHash := state.IntermediateRoot(false)
	state.Database().TrieDB().Commit(rootHash, false, nil)

	admServer := NewAdamniteServer(state, nil)
	admServer.Launch(nil)

	returnedInt := BigIntRPC{}
	admServer.GetBalance(testAddress, &returnedInt)
	assert.Equal(t, testBalance, returnedInt.toBigInt())
}
