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
	vm := newVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

	// The hash passed here should be the function index
	var getCodeMock = func(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
		var index = 0

		if hash[0] == 0x1 {
			index = 1
		}

		if hash[0] == 0x2 {
			index = 2
		}

		code, ctrlStack := parseBytes(vm.module.codeSection[index].body)
		return *vm.module.typeSection[0], code, ctrlStack
	}

	callCode := "00ee919d410a4102" // 0x00ee919d = FuncIdentifier, [0x41 = i32, value = 0x2] [0x41 = i64, value = 0x0a]
	vm.call2(callCode, getCodeMock)
	assert.Equal(t, vm.popFromStack(), uint64(0xc))

	vm.pointInCode = 0
	callCode2 := "01ee919d410a4102" // 0x01ee919d = FuncIdentifier, [0x41 = i32, value = 0x2] [0x41 = i64, value = 0x0a]

	vm.call2(callCode2, getCodeMock)

	assert.Equal(t, vm.popFromStack(), uint64(0x8))

	vm.pointInCode = 0
	callCode3 := "02ee919d410a4102" // 0x02ee919d = FuncIdentifier, [0x41 = i32, value = 0x2] [0x41 = i64, value = 0x0a]

	vm.call2(callCode3, getCodeMock)

	assert.Equal(t, vm.popFromStack(), uint64(0x14))
}
