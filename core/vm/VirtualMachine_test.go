package vm

import (
	"encoding/hex"
	"fmt"
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
	vm.config.codeGetter = getCodeMock

	callCode := "00ee919d00ee919d00ee919d00ee919d410a4102" // 0x00ee919d = FuncIdentifier, [0x41 = i32, value = 0x2] [0x41 = i64, value = 0x0a]
	vm.call2(callCode)
	assert.Equal(t, vm.popFromStack(), uint64(0xc))

	vm.pointInCode = 0
	callCode2 := "01ee919d01ee919d01ee919d01ee919d410a4102" // 0x01ee919d = FuncIdentifier, [0x41 = i32, value = 0x2] [0x41 = i64, value = 0x0a]

	vm.call2(callCode2)

	assert.Equal(t, vm.popFromStack(), uint64(0x8))

	vm.pointInCode = 0
	callCode3 := "02ee919d02ee919d02ee919d02ee919d410a4102" // 0x02ee919d = FuncIdentifier, [0x41 = i32, value = 0x2] [0x41 = i64, value = 0x0a]

	vm.call2(callCode3)

	assert.Equal(t, vm.popFromStack(), uint64(0x14))
}

func TestDataSection(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f0382808080000100048480808000017000000583808080000100010681808080000007918080800002066d656d6f72790200047465737400000a8a808080000184808080000041100b0b9280808000010041100b0c48656c6c6f20576f726c6400")
	vm := newVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

	module := *decode(wasmBytes)

	offset := module.dataSection[0].offsetExpression.data[0]
	size := len(module.dataSection[0].init) 

	// Read the data from data section
	
	s := string(vm.vmMemory[offset: size + int(offset)])
	
	fmt.Printf("s: %v\n", s)
	assert.Equal(t, "Hello World\x00", s)
}