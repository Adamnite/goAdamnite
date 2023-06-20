package VM

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_i32Add(t *testing.T) {

	wasmBytes, _ := hex.DecodeString("0061736d01000000010c0260027f7f017f60017f017f03201f0000000000000000000000000000000101010101010000000000000000000007de011f036164640000037375620001036d756c0002056469765f730003056469765f7500040572656d5f7300050572656d5f75000603616e640007026f72000803786f7200090373686c000a057368725f73000b057368725f75000c04726f746c000d04726f7472000e03636c7a000f0363747a001006706f70636e74001109657874656e64385f7300120a657874656e6431365f7300130365717a00140265710015026e650016046c745f730017046c745f750018046c655f730019046c655f75001a0467745f73001b0467745f75001c0467655f73001d0467655f75001e0aed011f0700200020016a0b0700200020016b0b0700200020016c0b0700200020016d0b0700200020016e0b0700200020016f0b070020002001700b070020002001710b070020002001720b070020002001730b070020002001740b070020002001750b070020002001760b070020002001770b070020002001780b05002000670b05002000680b05002000690b05002000c00b05002000c10b05002000450b070020002001460b070020002001470b070020002001480b070020002001490b0700200020014c0b0700200020014d0b0700200020014a0b0700200020014b0b0700200020014e0b0700200020014f0b00ef01046e616d6502e7011f00020001780101790102000178010179020200017801017903020001780101790402000178010179050200017801017906020001780101790702000178010179080200017801017909020001780101790a020001780101790b020001780101790c020001780101790d020001780101790e020001780101790f0100017810010001781101000178120100017813010001781401000178150200017801017916020001780101791702000178010179180200017801017919020001780101791a020001780101791b020001780101791c020001780101791d020001780101791e02000178010179")

	vm := NewVirtualMachine(wasmBytes, []uint64{}, 1000)

	module := *decode(wasmBytes)

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[0].body)
	vm.callStack[vm.currentFrame].Code = vm.vmCode

	// (assert_return (invoke "add" (i32.const 1) (i32.const 1)) (i32.const 2))
	vm.locals = append(vm.locals, 1)
	vm.locals = append(vm.locals, 1)
	vm.callStack[0].Locals = vm.locals
	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(2))
	// (assert_return (invoke "add" (i32.const 1) (i32.const 0)) (i32.const 1))

	vm.pointInCode = 0

	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 1)
	vm.locals = append(vm.locals, 0)

	// Reset main frame info since we're reusing the same
	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals
	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(1))

	// (assert_return (invoke "add" (i32.const -1) (i32.const -1)) (i32.const -2))
	// vm.locals = append(vm.locals, -1)
	// vm.locals = append(vm.locals, -1)
	// vm.run()
	// assert.Equal(t, vm.popFromStack(), -2)

	// (assert_return (invoke "add" (i32.const -1) (i32.const 1)) (i32.const 0))
	// vm.locals = append(vm.locals, -1)
	// vm.locals = append(vm.locals, 0)
	// vm.run()
	// assert.Equal(t, vm.popFromStack(), 0)

	// (assert_return (invoke "add" (i32.const 0x7fffffff) (i32.const 1)) (i32.const 0x80000000))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x7fffffff)
	vm.locals = append(vm.locals, 1)
	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals
	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x80000000))

	// (assert_return (invoke "add" (i32.const 0x80000000) (i32.const -1)) (i32.const 0x7fffffff))
	// (assert_return (invoke "add" (i32.const 0x80000000) (i32.const 0x80000000)) (i32.const 0))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x80000000)
	vm.locals = append(vm.locals, 0x80000000)
	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals
	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x0))

	// (assert_return (invoke "add" (i32.const 0x3fffffff) (i32.const 1)) (i32.const 0x40000000))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x3fffffff)
	vm.locals = append(vm.locals, 1)
	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals
	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x40000000))
}

func Test_i32Sub(t *testing.T) {

	wasmBytes, _ := hex.DecodeString("0061736d01000000010c0260027f7f017f60017f017f03201f0000000000000000000000000000000101010101010000000000000000000007de011f036164640000037375620001036d756c0002056469765f730003056469765f7500040572656d5f7300050572656d5f75000603616e640007026f72000803786f7200090373686c000a057368725f73000b057368725f75000c04726f746c000d04726f7472000e03636c7a000f0363747a001006706f70636e74001109657874656e64385f7300120a657874656e6431365f7300130365717a00140265710015026e650016046c745f730017046c745f750018046c655f730019046c655f75001a0467745f73001b0467745f75001c0467655f73001d0467655f75001e0aed011f0700200020016a0b0700200020016b0b0700200020016c0b0700200020016d0b0700200020016e0b0700200020016f0b070020002001700b070020002001710b070020002001720b070020002001730b070020002001740b070020002001750b070020002001760b070020002001770b070020002001780b05002000670b05002000680b05002000690b05002000c00b05002000c10b05002000450b070020002001460b070020002001470b070020002001480b070020002001490b0700200020014c0b0700200020014d0b0700200020014a0b0700200020014b0b0700200020014e0b0700200020014f0b00ef01046e616d6502e7011f00020001780101790102000178010179020200017801017903020001780101790402000178010179050200017801017906020001780101790702000178010179080200017801017909020001780101790a020001780101790b020001780101790c020001780101790d020001780101790e020001780101790f0100017810010001781101000178120100017813010001781401000178150200017801017916020001780101791702000178010179180200017801017919020001780101791a020001780101791b020001780101791c020001780101791d020001780101791e02000178010179")

	vm := NewVirtualMachine(wasmBytes, []uint64{}, 10000)
	vm.callStack[vm.currentFrame].Code = vm.vmCode

	module := *decode(wasmBytes)

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[1].body)
	vm.callStack[0].Code = vm.vmCode
	vm.callStack[0].CtrlStack = vm.controlBlockStack

	// (assert_return (invoke "sub" (i32.const 1) (i32.const 1)) (i32.const 0))
	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 1)
	vm.locals = append(vm.locals, 1)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0))

	// (assert_return (invoke "sub" (i32.const 1) (i32.const 0)) (i32.const 1))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 1)
	vm.locals = append(vm.locals, 0)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(1))

	// (assert_return (invoke "sub" (i32.const -1) (i32.const -1)) (i32.const 0))
	// (assert_return (invoke "sub" (i32.const 0x7fffffff) (i32.const -1)) (i32.const 0x80000000))

	// (assert_return (invoke "sub" (i32.const 0x80000000) (i32.const 1)) (i32.const 0x7fffffff))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x80000000)
	vm.locals = append(vm.locals, 1)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x7fffffff))

	// (assert_return (invoke "sub" (i32.const 0x80000000) (i32.const 0x80000000)) (i32.const 0))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x80000000)
	vm.locals = append(vm.locals, 0x80000000)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x0))

	// (assert_return (invoke "sub" (i32.const 0x3fffffff) (i32.const -1)) (i32.const 0x40000000))
}

func Test_i32divu(t *testing.T) {

	wasmBytes, _ := hex.DecodeString("0061736d01000000010c0260027f7f017f60017f017f03201f0000000000000000000000000000000101010101010000000000000000000007de011f036164640000037375620001036d756c0002056469765f730003056469765f7500040572656d5f7300050572656d5f75000603616e640007026f72000803786f7200090373686c000a057368725f73000b057368725f75000c04726f746c000d04726f7472000e03636c7a000f0363747a001006706f70636e74001109657874656e64385f7300120a657874656e6431365f7300130365717a00140265710015026e650016046c745f730017046c745f750018046c655f730019046c655f75001a0467745f73001b0467745f75001c0467655f73001d0467655f75001e0aed011f0700200020016a0b0700200020016b0b0700200020016c0b0700200020016d0b0700200020016e0b0700200020016f0b070020002001700b070020002001710b070020002001720b070020002001730b070020002001740b070020002001750b070020002001760b070020002001770b070020002001780b05002000670b05002000680b05002000690b05002000c00b05002000c10b05002000450b070020002001460b070020002001470b070020002001480b070020002001490b0700200020014c0b0700200020014d0b0700200020014a0b0700200020014b0b0700200020014e0b0700200020014f0b00ef01046e616d6502e7011f00020001780101790102000178010179020200017801017903020001780101790402000178010179050200017801017906020001780101790702000178010179080200017801017909020001780101790a020001780101790b020001780101790c020001780101790d020001780101790e020001780101790f0100017810010001781101000178120100017813010001781401000178150200017801017916020001780101791702000178010179180200017801017919020001780101791a020001780101791b020001780101791c020001780101791d020001780101791e02000178010179")

	vm := NewVirtualMachine(wasmBytes, []uint64{}, 100)
	vm.callStack[vm.currentFrame].Code = vm.vmCode

	module := *decode(wasmBytes)

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[4].body)
	vm.callStack[0].Code = vm.vmCode
	vm.callStack[0].CtrlStack = vm.controlBlockStack

	// (assert_trap (invoke "div_u" (i32.const 1) (i32.const 0)) "integer divide by zero")
	vm.locals = append(vm.locals, 1)
	vm.locals = append(vm.locals, 0)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	defer func() {
		if err := recover(); err != nil {
			assert.Equal(t, err, "Division by zero")
		}
	}()
	vm.run()

	// (assert_trap (invoke "div_u" (i32.const 0) (i32.const 0)) "integer divide by zero")
	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0)
	vm.locals = append(vm.locals, 0)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	defer func() {
		if err := recover(); err != nil {
			assert.Equal(t, err, "Division by zero")
		}
	}()
	vm.run()

	// (assert_return (invoke "div_u" (i32.const 1) (i32.const 1)) (i32.const 1))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 1)
	vm.locals = append(vm.locals, 1)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x1))

	// (assert_return (invoke "div_u" (i32.const 0) (i32.const 1)) (i32.const 0))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0)
	vm.locals = append(vm.locals, 1)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x1))

	// (assert_return (invoke "div_u" (i32.const -1) (i32.const -1)) (i32.const 1))
	// (assert_return (invoke "div_u" (i32.const 0x80000000) (i32.const -1)) (i32.const 0))
	// (assert_return (invoke "div_u" (i32.const 0x80000000) (i32.const 2)) (i32.const 0x40000000))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x80000000)
	vm.locals = append(vm.locals, 2)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x40000000))

	// (assert_return (invoke "div_u" (i32.const 0x8ff00ff0) (i32.const 0x10001)) (i32.const 0x8fef))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x8ff00ff0)
	vm.locals = append(vm.locals, 0x10001)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x8fef))

	// (assert_return (invoke "div_u" (i32.const 0x80000001) (i32.const 1000)) (i32.const 0x20c49b))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x80000001)
	vm.locals = append(vm.locals, 1000)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x20c49b))

	// (assert_return (invoke "div_u" (i32.const 5) (i32.const 2)) (i32.const 2))
	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x80000001)
	vm.locals = append(vm.locals, 1000)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x20c49b))

	// (assert_return (invoke "div_u" (i32.const -5) (i32.const 2)) (i32.const 0x7ffffffd))
	// (assert_return (invoke "div_u" (i32.const 5) (i32.const -2)) (i32.const 0))
	// (assert_return (invoke "div_u" (i32.const -5) (i32.const -2)) (i32.const 0))
	// (assert_return (invoke "div_u" (i32.const 7) (i32.const 3)) (i32.const 2))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 7)
	vm.locals = append(vm.locals, 3)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(2))

	// (assert_return (invoke "div_u" (i32.const 11) (i32.const 5)) (i32.const 2))
	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 11)
	vm.locals = append(vm.locals, 5)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(2))

	// (assert_return (invoke "div_u" (i32.const 17) (i32.const 7)) (i32.const 2))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 17)
	vm.locals = append(vm.locals, 7)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(2))

}
