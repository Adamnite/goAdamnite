package vm

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testAddress = []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11, 0x12, 0x13}
)

func TestOpAddress(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f0382808080000100048480808000017000000583808080000100010681808080000007918080800002066d656d6f72790200046d61696e00000aa280808000019c8080800001017f410028020441106b2200410036020c2000410a360208410a0b")
	vm := newVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	vm.chainHandler = BCSpoofer{contractAddress: testAddress}
	testCode := []byte{0xc1}
	module := *decode(wasmBytes)
	foo := module.codeSection[0].body
	for i := range foo {
		testCode = append(testCode, foo[i])
	}
	vm.vmCode, vm.controlBlockStack = parseBytes(testCode)
	vm.step()
	assert.Equal(t, testAddress, uintsArrayToAddress(vm.vmStack))
	//yes, this is a horribly lazy way to test our custom opcodes, and i should write the functions correctly... Or at least use one that breaks things less...
}
