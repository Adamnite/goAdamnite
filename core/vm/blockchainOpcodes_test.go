package vm

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testAddress = []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11, 0x12, 0x13}
)

func TestOpAddress(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d0100000001070160027f7f017f03020100070a010661646454776f00000a09010700200020016a0b000a046e616d650203010000")
	vm := newVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	spoofer := newBCSpoofer()
	spoofer.contractAddress = testAddress

	testCode := []byte{Op_address}
	module := *decode(wasmBytes)
	foo := module.codeSection[0].body
	for i := range foo {
		testCode = append(testCode, foo[i])
	}
	vm.vmCode, vm.controlBlockStack = parseBytes(testCode)
	vm.step()
	assert.Equal(t, testAddress, uintsArrayToAddress(vm.vmStack))
	//yes, this is a horribly lazy way to test our custom opcodes, and i should write the functions correctly...
}

func TestOpBalance(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d0100000001070160027f7f017f03020100070a010661646454776f00000a09010700200020016a0b000a046e616d650203010000")
	vm := newVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	testBalance := big.NewInt(9000000000000000000)
	testBalance.Mul(testBalance, big.NewInt(100))

	// fmt.Println(big.NewInt(0).SetBytes(uintsArrayToAddress(addressToInts(testBalance.Bytes()))))
	spoofer := newBCSpoofer()
	spoofer.contractAddress = testAddress
	spoofer.setBalanceFromByteAddress(testAddress, *testBalance)

	testCode := []byte{
		Op_address,
		Op_balance,
	}
	module := *decode(wasmBytes)

	foo := module.codeSection[0].body
	for i := range foo {
		testCode = append(testCode, foo[i])
	}
	vm.vmCode, vm.controlBlockStack = parseBytes(testCode)
	vm.step()
	vm.step()
	assert.Equal(t, testBalance, arrayToBalance(vm.vmStack))
	//yes, this is a horribly lazy way to test our custom opcodes, and i should write the functions correctly...
}
