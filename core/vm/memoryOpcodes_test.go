package vm

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)


func Test_i32Store(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f0382808080000100048480808000017000000583808080000100010681808080000007918080800002066d656d6f72790200046d61696e00000aa280808000019c8080800001017f410028020441106b2200410036020c2000410a360208410a0b")

	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})
	vm.debugStack = true
	expectedModuleCode := []byte{
		Op_i32_const, 0x0, 
		Op_i32_load, 0x2, 0x4, 
		Op_i32_const, 0x10, 
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
	
	vm.run()
	// Stored value in memory
	r := LE.Uint32(vm.vmMemory[10 : 14])
	// Return value
	assert.Equal(t, uint64(vm.popFromStack()), uint64(0xa)) 
	assert.Equal(t, uint64(r), uint64(0xa))
	assert.Equal(t, expectedModuleCode, vm.module.codeSection[0].body)
}


func Test_i32Store2(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f0382808080000100048480808000017000000583808080000100010681808080000007918080800002066d656d6f72790200046d61696e00000aa98080800001a38080800001017f410028020441106b2200410036020c2000410a3602082000411936020441190b")

	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})
	vm.debugStack = true
	expectedModuleCode := []byte{
		Op_i32_const, 0x0, 
		Op_i32_load, 0x2, 0x4, 
		Op_i32_const, 0x10, 
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
	
	vm.run()
	// Stored value in memory
	r := LE.Uint32(vm.vmMemory[10 : 14])
	// Return value
	assert.Equal(t, uint64(vm.popFromStack()), uint64(25)) 
	assert.Equal(t, uint64(r), uint64(0xa))
	assert.Equal(t, expectedModuleCode, vm.module.codeSection[0].body)
}