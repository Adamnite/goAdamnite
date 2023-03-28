package main

import (
	"fmt"
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/params"
	"github.com/adamnite/go-adamnite/rpc"
)

// this is entirely setup to run a adamnite RPC server on local host. Do not use for any further purpose.
var (
	niteBigExponent = big.NewInt(1).Exp(big.NewInt(10), big.NewInt(20), nil) //big math version of 10**20
	testAccounts    = []common.Address{                                      //contains address 0, 1, 2.
		common.Address{0},
		common.BytesToAddress([]byte{0x01}),
		common.BytesToAddress([]byte{0x02}),
	}
	testBalances = []*big.Int{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(2),
	}
)

func main() {
	test_db := rawdb.NewMemoryDB()
	state, _ := statedb.New(common.Hash{}, statedb.NewDatabase(test_db))
	chainConfig := params.TestnetChainConfig

	//converting testBalances from "nite" to dubnite. then adds the balance to a DB
	for i, foo := range testBalances {
		testBalances[i] = big.NewInt(0).Mul(niteBigExponent, foo)
		state.AddBalance(testAccounts[i], testBalances[i])
	}
	rootHash := state.IntermediateRoot(false)
	state.Database().TrieDB().Commit(rootHash, false, nil)

	bc, err := core.NewBlockchain(test_db,
		chainConfig,
		dpos.New(chainConfig, test_db),
	)
	if err != nil {
		fmt.Println(err)
	}

	as := rpc.NewAdamniteServer(state, bc)
	foo := "[127.0.0.1]:12345"
	as.Launch(&foo)
	time.Sleep(time.Minute * 5) //change this time for how long you want the server to stay up
	// for {}//uncomment to have run forever.
}
