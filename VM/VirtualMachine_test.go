package VM

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
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

	// The hash passed here should be the function index
	var getCodeMock = func(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
		var index = 0

		if hash[0] == 0x1 {
			index = 1
		}

		if hash[0] == 0x2 {
			index = 2
		}
		module := *decode(wasmBytes)
		code, ctrlStack := parseBytes(module.codeSection[index].body)
		return *module.typeSection[0], code, ctrlStack
	}
	vm.config.CodeGetter = getCodeMock

	callCode := "00ee919d00ee919d00ee919d00ee919d7f0a7f02"
	// 00ee919d00ee919d00ee919d00ee919d = FuncIdentifier, [0x7f = i32, value = 0x2] [0x7f = i64, value = 0x0a]
	vm.Call2(callCode, 10000)
	assert.Equal(t, vm.popFromStack(), uint64(0xc))

	vm.pointInCode = 0
	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	callCode2 := "01ee919d01ee919d01ee919d01ee919d7f0a7f02"
	// 01ee919d01ee919d01ee919d01ee919d = FuncIdentifier, [0x7f = i32, value = 0x2] [0x7f = i64, value = 0x0a]

	vm.Call2(callCode2, 1000)

	assert.Equal(t, vm.popFromStack(), uint64(0x8))

	vm.pointInCode = 0
	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	callCode3 := "02ee919d02ee919d02ee919d02ee919d7f0a7f02" // 0x02ee919d = FuncIdentifier, [0x7f = i32, value = 0x2] [0x7f = i64, value = 0x0a]

	vm.Call2(callCode3, 1000)

	assert.Equal(t, vm.popFromStack(), uint64(0x14))
}

func TestDataSection(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f0382808080000100048480808000017000000583808080000100010681808080000007918080800002066d656d6f72790200047465737400000a8a808080000184808080000041100b0b9280808000010041100b0c48656c6c6f20576f726c6400")
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	module := *decode(wasmBytes)
	initMemoryWithDataSection(&module, vm)

	offset := module.dataSection[0].offsetExpression.data[0]
	size := len(module.dataSection[0].init)

	// Read the data from data section

	s := string(vm.vmMemory[offset : size+int(offset)])

	fmt.Printf("s: %v\n", s)
	assert.Equal(t, "Hello World\x00", s)
}

func TestMultiDataSection(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f0382808080000100048480808000017000000583808080000100010681808080000007918080800002066d656d6f72790200046d61696e00000aa98080800001a38080800001017f410028020441106b2200410036020c2000411036020820004120360204410a0b0b9780808000020041100b0648656c6c6f000041200b06576f726c6400")
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

	module := *decode(wasmBytes)
	initMemoryWithDataSection(&module, vm)

	offset := module.dataSection[0].offsetExpression.data[0]
	size := len(module.dataSection[0].init)

	offset2 := module.dataSection[1].offsetExpression.data[0]
	size2 := len(module.dataSection[1].init)

	// Read the data from data section

	s := string(vm.vmMemory[offset : size+int(offset)])
	s += string(vm.vmMemory[offset2 : size2+int(offset2)])

	fmt.Printf("s: %v\n", s)
	assert.Equal(t, "Hello\x00World\x00", s)
}

func TestGettingFinalDataChanges(t *testing.T) {
	vm := NewVirtualMachine([]byte{}, []uint64{}, nil, 1000)
	const (
		runCount          = 500
		largeNumberOffset = 1100
	)
	// vm.config.debugStack = true
	vm.callStack[0].Code = []OperationCommon{
		localGet{0, 0},
		localGet{0, 0},
		GlobalSet{-1, 0},
		i64Const{0xFFFF, 0}, //FFFF stands out pretty easily to check for
		localGet{0, 0},
		i64Const{largeNumberOffset, 0},
		i64Mul{0},
		GlobalSet{-1, 0},
	}
	for i := 0; i < runCount; i++ {
		vm.currentFrame = 0
		vm.callStack[0].Locals = []uint64{uint64(i)}
		vm.callStack[0].Ip = 0
		vm.pointInCode = 0
		vm.run()
	}
	changes := vm.GetChanges()

	//generate the version of the changes we *should* see
	goalTotalChanges := runCount //runcount-1 for the *10 numbers, then 1 for the main
	firstPointChanges := []byte{0xff, 0xFF, 0, 0, 0, 0, 0, 0}
	for i := 1; i < runCount; i++ {
		LE.PutUint64(firstPointChanges, uint64(i))
	}
	assert.Equal(t,
		goalTotalChanges,
		len(changes.ChangeStartPoints),
		"incorrect change count",
	)
	assert.Equal(t,
		firstPointChanges,
		changes.Changed[0],
		"changes do not match theoretical",
	)
	for i, v := range changes.ChangeStartPoints {
		if i == 0 {
			assert.Equal(t,
				uint64(0),
				v,
				"incorrect start",
			)
		} else {
			assert.Equal(t,
				uint64((i)*largeNumberOffset*8),
				v,
				"incorrect start on later values",
			)
		}

	}
	// changes.OutputChanges()
}
