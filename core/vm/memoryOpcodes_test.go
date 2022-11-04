package vm

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_i32Store(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f0382808080000100048480808000017000000583808080000100010681808080000007918080800002066d656d6f72790200046d61696e00000aa280808000019c8080800001017f410028020441106b2200410036020c2000410a360208410a0b")

	vm := newVirtualMachine(wasmBytes, []byte{}, VMConfig{})
	vm.debugStack = true
	code := []byte{
		Op_i32_const, 0x0,
		Op_i32_const, 0x10,
		Op_i32_load, 0x2, 0x4,
		Op_i32_sub,
		Op_tee_local, 0x0,
		Op_i32_const, 0x0,
		Op_i32_store, 0x2, 0xc,
		Op_get_local, 0x0,
		Op_i32_const, 0xa,
		Op_i32_store, 0x2, 0x8,
		Op_i32_const, 0xa,
		Op_end,
	}
	vm.vmCode, vm.controlBlockStack = parseBytes(code)
	vm.run()
	// Stored values in memory
	stored1 := LE.Uint32(vm.vmMemory[12 : 12+4])
	stored2 := LE.Uint32(vm.vmMemory[0x8 : 0x8+4])
	assert.Equal(t, uint64(0x0), uint64(stored1))
	assert.Equal(t, uint64(0xa), uint64(stored2))

	// Return value
	assert.Equal(t, uint64(vm.popFromStack()), uint64(0xa))
}

func Test_i32Store2(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f0382808080000100048480808000017000000583808080000100010681808080000007918080800002066d656d6f72790200046d61696e00000aa98080800001a38080800001017f410028020441106b2200410036020c2000410a3602082000411936020441190b")

	vm := newVirtualMachine(wasmBytes, []byte{}, VMConfig{})
	vm.debugStack = true
	code := []byte{
		Op_i32_const, 0x0, 
		Op_i32_const, 0x10, 
		Op_i32_load, 0x2, 0x4, 
		Op_i32_sub, 
		Op_tee_local, 0x0, 
		Op_i32_const, 0x0, 
		Op_i32_store, 0x2, 0xc, 
		Op_get_local, 0x0, 
		Op_i32_const, 0xa, 
		Op_i32_store, 0x2, 0x8, 
		Op_get_local, 0x0, 
		Op_i32_const, 0x19, 
		Op_i32_store, 0x2, 0x4, 
		Op_i32_const, 0x19,
		Op_end,
	}
	vm.vmCode, vm.controlBlockStack = parseBytes(code)

	vm.run()
	// Stored values in memory
	stored1 := LE.Uint32(vm.vmMemory[0xc : 0xc+4])
	stored2 := LE.Uint32(vm.vmMemory[0x8 : 0x8+4])
	stored3 := LE.Uint32(vm.vmMemory[0x4 : 0x4+4])
	assert.Equal(t, uint64(0x0), uint64(stored1))
	assert.Equal(t, uint64(0xa), uint64(stored2))
	assert.Equal(t, uint64(0x19), uint64(stored3))

	// Return value
	assert.Equal(t, uint64(vm.popFromStack()), uint64(0x19))
}

func Test_i32Store3(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f03828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000a978080800001918080800000410028020441106b410436020800000b")

	vm := newVirtualMachine(wasmBytes, []byte{}, VMConfig{})
	vm.debugStack = true
	code := []byte{
		Op_i32_const, 0x0,
		Op_i32_const, 0x10,
		Op_i32_load, 0x2, 0x4,
		Op_i32_sub,
		Op_i32_const, 0x4,
		Op_i32_store, 0x2, 0x8,
		0x0, 0x0,
		Op_end,
	}

	vm.vmCode, vm.controlBlockStack = parseBytes(code)
	vm.run()
	// Stored value in memory
	r := LE.Uint32(vm.vmMemory[0x8 : 0x8+4])
	assert.Equal(t, uint64(0x4), uint64(r))
}
