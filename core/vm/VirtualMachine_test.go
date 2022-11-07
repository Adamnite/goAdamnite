package vm

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestCall2(t *testing.T) {
	// Our module code
	// (module
	// 	(func $add (param i32 i32) (result i32)
	// 	  local.get 0
	// 	  local.get 1
	// 	  i32.add)
		
	// 	(func $sub (param i32 i32) (result i32)
	// 	  local.get 0
	// 	  local.get 1
	// 	  i32.sub)
		
	// 	(func $mul (param i32 i32) (result i32)
	// 	  local.get 0
	// 	  local.get 1
	// 	  i32.mul)
	//   )
		  
	wasmBytes, _ := hex.DecodeString("0061736d0100000001070160027f7f017f0304030000000a19030700200020016a0b0700200020016b0b0700200020016c0b0020046e616d650110030003616464010373756202036d756c020703000001000200")
	vm := newVirtualMachine(wasmBytes, []byte{}, VMConfig{})

	callCode := "01410a4102" // 0x01 = FuncIndex, [0x41 = i32, value = 0x2] [0x41 = i64, value = 0x0a]

	vm.call2(callCode)

	assert.Equal(t, vm.popFromStack(), uint64(0x8))

	vm.pointInCode = 0
	callCode2 := "00410a4102" // 0x00 = FuncIndex, [0x41 = i32, value = 0x2] [0x41 = i64, value = 0x0a]

	vm.call2(callCode2)

	assert.Equal(t, vm.popFromStack(), uint64(0xc))
}