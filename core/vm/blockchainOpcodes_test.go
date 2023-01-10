package vm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/stretchr/testify/assert"
)

var (
	testAddress = []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11, 0x12, 0x13}
	spoofer     DBSpoofer
	vm          *Machine
	hashes      []string
)

func preTestSetup() {
	getContractAddressWasm := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x07, 0x01, 0x60,
		0x00, 0x04, 0x7e, 0x7e, 0x7e, 0x7e, 0x03, 0x02, 0x01, 0x00, 0x07, 0x0e, 0x01,
		0x0a, 0x67, 0x65, 0x74, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x00,
		0x00, 0x0a, 0x0a, 0x01, 0x03, 0x00, Op_address,
		0x0b,
	}
	// (module
	// 	(type (;0;) (func (result i64 i64 i64)))
	// 	(func (;0;) (type 0) (result i64 i64 i64)
	// 		contractAddress
	// 	)
	// 	(export "getBalance" (func 0)))
	getContractBalanceWasm := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x07, 0x01, 0x60,
		0x00, 0x04, 0x7e, 0x7e, 0x7e, 0x7e, 0x03, 0x02, 0x01, 0x00, 0x07, 0x0e, 0x01,
		0x0a, 0x67, 0x65, 0x74, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x00,
		0x00, 0x0a, 0x0a, 0x01, 0x04, 0x00, Op_address, Op_balance,
		0x0b,
	}
	// (module
	// 	(type (;0;) (func (result i64 i64 i64)))
	// 	(func (;0;) (type 0) (result i64 i64 i64)
	// 		contractAddress
	// 		balance
	// 	)
	// 	(export "getBalance" (func 0)))
	spoofer = newDBSpoofer()
	err, foo := spoofer.addModuleToSpoofedCode([][]byte{getContractAddressWasm, getContractBalanceWasm})
	for i := 0; i < len(foo); i++ {
		hashes = append(hashes, hex.EncodeToString(foo[i]))
	}
	if err != nil {
		panic("error in preTestSetup")
	}
	vm = newVirtualMachine([]byte(emptyModule()), []uint64{}, nil, 1000)
	vm.contract.Address = common.BytesToAddress(testAddress)
	vm.config.codeGetter = spoofer.GetCode
}

func TestOpAddress(t *testing.T) {
	preTestSetup()
	vm.config.debugStack = true
	fmt.Println(vm.call2(hashes[0]+"", 1000))

	fmt.Println(vm.outputStack())
	assert.Equal(t, vm.contract.Address.Bytes(), uintsArrayToAddress(vm.vmStack))
}

func TestOpBalance(t *testing.T) {
	preTestSetup()
	testBalance := big.NewInt(9000000000000000000)
	testBalance.Mul(testBalance, big.NewInt(100))

	db := rawdb.NewMemoryDB()
	state, _ := statedb.New(common.Hash{}, statedb.NewDatabase(db))
	state.AddBalance(common.BytesToAddress(testAddress), testBalance)
	rootHash := state.IntermediateRoot(false)
	state.Database().TrieDB().Commit(rootHash, false, nil)
	vm.Statedb = state

	// fmt.Println(state.GetBalance(common.BytesToAddress(testAddress)))
	fmt.Println(vm.call2(hashes[1]+"", 1000))
	assert.Equal(t, testBalance, arrayToBalance(vm.vmStack))
}
