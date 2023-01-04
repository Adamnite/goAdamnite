package vm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

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
		0x00, 0x03, 0x7e, 0x7e, 0x7e, 0x03, 0x02, 0x01, 0x00, 0x07, 0x0e, 0x01,
		0x0a, 0x67, 0x65, 0x74, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x00,
		// 0x00, 0x0a, 0x0a, 0x01, 0x08, 0x00, 0x42, 0x01, 0x42, 0x01, 0x41, 0x01,
		0x00, 0x0a, 0x0a, 0x01, 0x03, 0x00, Op_address,//either have this or the line above uncommented
		0x0b,
	}
	// (module
	// 	(type (;0;) (func (result i64 i64 i64)))
	// 	(func (;0;) (type 0) (result i64 i64 i64)
	// 		contractAddress
	// 	)
	// 	(export "getBalance" (func 0)))
	spoofer = newDBSpoofer()
	module := *decode(getContractAddressWasm)
	err, foo := spoofer.addModuleToSpoofedCode(module)
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

	assert.Equal(t, testAddress, uintsArrayToAddress(vm.vmStack))
}

func TestOpBalance(t *testing.T) {
	wasmBytes := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x07, 0x01, 0x60,
		0x00, 0x03, 0x7e, 0x7e, 0x7e, 0x03, 0x02, 0x01, 0x00, 0x07, 0x0e, 0x01,
		0x0a, 0x67, 0x65, 0x74, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x00,
		0x00, 0x0a, 0x0a, 0x01, 0x08, 0x00, 0xc1, 0xc2, //0xc1 is address of contract 0xc2 is get Balance
		0x0b, 0x00, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x02, 0x03, 0x01, 0x00,
		0x00,
	}
	// (module
	// 	(type (;0;) (func (result i64 i64 i64)))
	// 	(func (;0;) (type 0) (result i64 i64 i64)
	// 		contractAddress
	// 		balance
	// 	)
	// 	(export "getBalance" (func 0)))
	vm := newVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	testBalance := big.NewInt(9000000000000000000)
	testBalance.Mul(testBalance, big.NewInt(100))

	// fmt.Println(big.NewInt(0).SetBytes(uintsArrayToAddress(addressToInts(testBalance.Bytes()))))
	spoofer := newBCSpoofer()
	spoofer.contractAddress = testAddress
	spoofer.setBalanceFromByteAddress(testAddress, *testBalance)

	// testCode := []byte{
	// 	Op_address,
	// 	Op_balance,
	// }
	// module := *decode(wasmBytes)

	// foo := module.codeSection[0].body
	// for i := range foo {
	// 	testCode = append(testCode, foo[i])
	// }
	// module := *decode(wasmBytes)

	vm.vmCode, vm.controlBlockStack = parseBytes(wasmBytes)
	vm.run()
	assert.Equal(t, testBalance, arrayToBalance(vm.vmStack))
	//yes, this is a horribly lazy way to test our custom opcodes, and i should write the functions correctly...
}
