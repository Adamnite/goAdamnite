package VM

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_f32Basics(t *testing.T) {
	wasmBytes := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x07, 0x01, 0x60,
		0x02, 0x7d, 0x7d, 0x01, 0x7d, 0x03, 0x05, 0x04, 0x00, 0x00, 0x00, 0x00,
		0x07, 0x29, 0x04, 0x07, 0x66, 0x61, 0x64, 0x64, 0x54, 0x77, 0x6f, 0x00,
		0x00, 0x07, 0x66, 0x73, 0x75, 0x62, 0x54, 0x77, 0x6f, 0x00, 0x01, 0x07,
		0x66, 0x6d, 0x75, 0x6c, 0x54, 0x77, 0x6f, 0x00, 0x02, 0x07, 0x66, 0x64,
		0x69, 0x76, 0x54, 0x77, 0x6f, 0x00, 0x03, 0x0a, 0x21, 0x04, 0x07, 0x00,
		0x20, 0x00, 0x20, 0x01, 0x92, 0x0b, 0x07, 0x00, 0x20, 0x00, 0x20, 0x01,
		0x93, 0x0b, 0x07, 0x00, 0x20, 0x00, 0x20, 0x01, 0x94, 0x0b, 0x07, 0x00,
		0x20, 0x00, 0x20, 0x01, 0x95, 0x0b, 0x00, 0x10, 0x04, 0x6e, 0x61, 0x6d,
		0x65, 0x02, 0x09, 0x04, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00,
	}
	//   (module
	// 	(type (;0;) (func (param f32 f32) (result f32)))
	// 	(func (;0;) (type 0) (param f32 f32) (result f32)
	// 	  local.get 0
	// 	  local.get 1
	// 	  f32.add)
	// 	(func (;1;) (type 0) (param f32 f32) (result f32)
	// 	  local.get 0
	// 	  local.get 1
	// 	  f32.sub)
	// 	(func (;2;) (type 0) (param f32 f32) (result f32)
	// 	  local.get 0
	// 	  local.get 1
	// 	  f32.mul)
	// 	(func (;3;) (type 0) (param f32 f32) (result f32)
	// 	  local.get 0
	// 	  local.get 1
	// 	  f32.div)
	// 	(export "faddTwo" (func 0))
	// 	(export "fsubTwo" (func 1))
	// 	(export "fmulTwo" (func 2))
	// 	(export "fdivTwo" (func 3)))

	// testCode := []byte{}
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	// vm.config.debugStack = true

	module := *decode(wasmBytes)

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[0].body)
	testParams := [][]float32{
		{10, 100000},
		{10, -100},
	}
	for i := range testParams { //f32_add
		vm.Reset()
		vm.AddLocal(testParams[i])
		vm.callStack[0].Locals = vm.locals
		vm.run()
		poppedValue := vm.popFromStack()
		assert.Equal(t, testParams[i][0]+testParams[i][1], math.Float32frombits(uint32(poppedValue)))

	}

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[1].body)
	for i := range testParams { //f32_sub
		vm.Reset()
		vm.AddLocal(testParams[i])
		vm.callStack[0].Locals = vm.locals
		vm.run()
		poppedValue := vm.popFromStack()
		assert.Equal(t, testParams[i][0]-testParams[i][1], math.Float32frombits(uint32(poppedValue)))

	}

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[2].body)
	for i := range testParams { //f32_mul
		vm.Reset()
		vm.AddLocal(testParams[i])
		vm.callStack[0].Locals = vm.locals
		vm.run()
		poppedValue := vm.popFromStack()
		assert.Equal(t, testParams[i][0]*testParams[i][1], math.Float32frombits(uint32(poppedValue)))

	}

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[3].body)
	for i := range testParams { //f32_div
		vm.Reset()
		vm.AddLocal(testParams[i])
		vm.callStack[0].Locals = vm.locals
		vm.run()
		poppedValue := vm.popFromStack()
		assert.Equal(t, testParams[i][0]/testParams[i][1], math.Float32frombits(uint32(poppedValue)))

	}

}

func Test_f64Basics(t *testing.T) {
	wasmBytes := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x07, 0x01, 0x60,
		0x02, 0x7c, 0x7c, 0x01, 0x7c, 0x03, 0x05, 0x04, 0x00, 0x00, 0x00, 0x00,
		0x07, 0x29, 0x04, 0x07, 0x66, 0x61, 0x64, 0x64, 0x54, 0x77, 0x6f, 0x00,
		0x00, 0x07, 0x66, 0x73, 0x75, 0x62, 0x54, 0x77, 0x6f, 0x00, 0x01, 0x07,
		0x66, 0x6d, 0x75, 0x6c, 0x54, 0x77, 0x6f, 0x00, 0x02, 0x07, 0x66, 0x64,
		0x69, 0x76, 0x54, 0x77, 0x6f, 0x00, 0x03, 0x0a, 0x21, 0x04, 0x07, 0x00,
		0x20, 0x00, 0x20, 0x01, 0xa0, 0x0b, 0x07, 0x00, 0x20, 0x00, 0x20, 0x01,
		0xa1, 0x0b, 0x07, 0x00, 0x20, 0x00, 0x20, 0x01, 0xa2, 0x0b, 0x07, 0x00,
		0x20, 0x00, 0x20, 0x01, 0xa3, 0x0b, 0x00, 0x10, 0x04, 0x6e, 0x61, 0x6d,
		0x65, 0x02, 0x09, 0x04, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00,
	}
	//   (module
	// 	(type (;0;) (func (param f64 f64) (result f64)))
	// 	(func (;0;) (type 0) (param f64 f64) (result f64)
	// 	  local.get 0
	// 	  local.get 1
	// 	  f64.add)
	// 	(func (;1;) (type 0) (param f64 f64) (result f64)
	// 	  local.get 0
	// 	  local.get 1
	// 	  f64.sub)
	// 	(func (;2;) (type 0) (param f64 f64) (result f64)
	// 	  local.get 0
	// 	  local.get 1
	// 	  f64.mul)
	// 	(func (;3;) (type 0) (param f64 f64) (result f64)
	// 	  local.get 0
	// 	  local.get 1
	// 	  f64.div)
	// 	(export "faddTwo" (func 0))
	// 	(export "fsubTwo" (func 1))
	// 	(export "fmulTwo" (func 2))
	// 	(export "fdivTwo" (func 3)))

	// testCode := []byte{}
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	vm.config.debugStack = true

	module := *decode(wasmBytes)

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[0].body)
	testParams := [][]float64{
		{10, 100000},
		{10, -100},
	}
	for i := range testParams { //f64_add
		vm.Reset()
		vm.AddLocal(testParams[i])
		vm.callStack[0].Locals = vm.locals
		vm.run()
		poppedValue := vm.popFromStack()
		assert.Equal(t, testParams[i][0]+testParams[i][1], math.Float64frombits(poppedValue))

	}

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[1].body)
	for i := range testParams { //f64_sub
		vm.Reset()
		vm.AddLocal(testParams[i])
		vm.callStack[0].Locals = vm.locals
		vm.run()
		poppedValue := vm.popFromStack()
		assert.Equal(t, testParams[i][0]-testParams[i][1], math.Float64frombits(poppedValue))

	}

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[2].body)
	for i := range testParams { //f64_mul
		vm.Reset()
		vm.AddLocal(testParams[i])
		vm.callStack[0].Locals = vm.locals
		vm.run()
		poppedValue := vm.popFromStack()
		assert.Equal(t, testParams[i][0]*testParams[i][1], math.Float64frombits(poppedValue))

	}

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[3].body)
	for i := range testParams { //f64_div
		vm.Reset()
		vm.AddLocal(testParams[i])
		vm.callStack[0].Locals = vm.locals
		vm.run()
		poppedValue := vm.popFromStack()
		assert.Equal(t, testParams[i][0]/testParams[i][1], math.Float64frombits(poppedValue))

	}

}
