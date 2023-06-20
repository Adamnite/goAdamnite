package VM

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testSetup(funcIndex int) {
	wasmBytes, _ := hex.DecodeString("0061736d0100000001170460027e7e017e60017e017e60017e017f60027e7e017f032120000000000000000000000000000000010101010101020303030303030303030307eb0120036164640000037375620001036d756c0002056469765f730003056469765f7500040572656d5f7300050572656d5f75000603616e640007026f72000803786f7200090373686c000a057368725f73000b057368725f75000c04726f746c000d04726f7472000e03636c7a000f0363747a001006706f70636e74001109657874656e64385f7300120a657874656e6431365f7300130a657874656e6433325f7300140365717a00150265710016026e650017046c745f730018046c745f750019046c655f73001a046c655f75001b0467745f73001c0467745f75001d0467655f73001e0467655f75001f0af301200700200020017c0b0700200020017d0b0700200020017e0b0700200020017f0b070020002001800b070020002001810b070020002001820b070020002001830b070020002001840b070020002001850b070020002001860b070020002001870b070020002001880b070020002001890b0700200020018a0b05002000790b050020007a0b050020007b0b05002000c20b05002000c30b05002000c40b05002000500b070020002001510b070020002001520b070020002001530b070020002001540b070020002001570b070020002001580b070020002001550b070020002001560b070020002001590b0700200020015a0b00f401046e616d6502ec012000020001780101790102000178010179020200017801017903020001780101790402000178010179050200017801017906020001780101790702000178010179080200017801017909020001780101790a020001780101790b020001780101790c020001780101790d020001780101790e020001780101790f0100017810010001781101000178120100017813010001781401000178150100017816020001780101791702000178010179180200017801017919020001780101791a020001780101791b020001780101791c020001780101791d020001780101791e020001780101791f02000178010179")

	vm = NewVirtualMachine(wasmBytes, []uint64{}, 1000)

	module := *decode(wasmBytes)

	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[funcIndex].body)
	vm.callStack[vm.currentFrame].Code = vm.vmCode
}
func Test_i64Add(t *testing.T) {

	testSetup(0)
	// (assert_return (invoke "add" (i64.const 1) (i64.const 1)) (i64.const 2))
	vm.locals = append(vm.locals, 1)
	vm.locals = append(vm.locals, 1)
	vm.callStack[0].Locals = vm.locals
	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(2))
	// (assert_return (invoke "add" (i64.const 1) (i64.const 0)) (i64.const 1))

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

	// (assert_return (invoke "add" (i64.const 0x7fffffffffffffff) (i64.const 1)) (i64.const 0x8000000000000000))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x7fffffffffffffff)
	vm.locals = append(vm.locals, 1)
	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals
	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x8000000000000000))

	// (assert_return (invoke "add" (i64.const 0x8000000000000000) (i64.const -1)) (i64.const 0x7fffffffffffffff))
	// (assert_return (invoke "add" (i64.const 0x8000000000000000) (i64.const 0x8000000000000000)) (i64.const 0))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x8000000000000000)
	vm.locals = append(vm.locals, 0x8000000000000000)
	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals
	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x0))

	// (assert_return (invoke "add" (i64.const 0x3fffffff) (i64.const 1)) (i64.const 0x40000000))

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

func Test_i64Sub(t *testing.T) {

	testSetup(1)
	// (assert_return (invoke "sub" (i64.const 1) (i64.const 1)) (i64.const 0))
	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 1)
	vm.locals = append(vm.locals, 1)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0))

	// (assert_return (invoke "sub" (i64.const 1) (i64.const 0)) (i64.const 1))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 1)
	vm.locals = append(vm.locals, 0)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(1))

	// (assert_return (invoke "sub" (i64.const -1) (i64.const -1)) (i64.const 0))
	// (assert_return (invoke "sub" (i64.const 0x7fffffffffffffff) (i64.const -1)) (i64.const 0x8000000000000000))

	// (assert_return (invoke "sub" (i64.const 0x8000000000000000) (i64.const 1)) (i64.const 0x7fffffffffffffff))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x8000000000000000)
	vm.locals = append(vm.locals, 1)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x7fffffffffffffff))

	// (assert_return (invoke "sub" (i64.const 0x8000000000000000) (i64.const 0x8000000000000000)) (i64.const 0))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x8000000000000000)
	vm.locals = append(vm.locals, 0x8000000000000000)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x0))

	// (assert_return (invoke "sub" (i64.const 0x3fffffff) (i64.const -1)) (i64.const 0x40000000))
}

func Test_i64divu(t *testing.T) {

	testSetup(4)
	// (assert_trap (invoke "div_u" (i64.const 1) (i64.const 0)) "integer divide by zero")
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

	// (assert_trap (invoke "div_u" (i64.const 0) (i64.const 0)) "integer divide by zero")
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

	// (assert_return (invoke "div_u" (i64.const 1) (i64.const 1)) (i64.const 1))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 1)
	vm.locals = append(vm.locals, 1)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x1))

	// (assert_return (invoke "div_u" (i64.const 0) (i64.const 1)) (i64.const 0))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0)
	vm.locals = append(vm.locals, 1)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x1))

	// (assert_return (invoke "div_u" (i64.const -1) (i64.const -1)) (i64.const 1))
	// (assert_return (invoke "div_u" (i64.const 0x8000000000000000) (i64.const -1)) (i64.const 0))
	// (assert_return (invoke "div_u" (i64.const 0x8000000000000000) (i64.const 2)) (i64.const 0x40000000))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x8000000000000000)
	vm.locals = append(vm.locals, 2)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x40000000))

	// (assert_return (invoke "div_u" (i64.const 0x8ff00ff0) (i64.const 0x10001)) (i64.const 0x8fef))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x8ff00ff0)
	vm.locals = append(vm.locals, 0x10001)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x8fef))

	// (assert_return (invoke "div_u" (i64.const 0x80000001) (i64.const 1000)) (i64.const 0x20c49b))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x80000001)
	vm.locals = append(vm.locals, 1000)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x20c49b))

	// (assert_return (invoke "div_u" (i64.const 5) (i64.const 2)) (i64.const 2))
	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 0x80000001)
	vm.locals = append(vm.locals, 1000)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0x20c49b))

	// (assert_return (invoke "div_u" (i64.const -5) (i64.const 2)) (i64.const 0x7ffffffd))
	// (assert_return (invoke "div_u" (i64.const 5) (i64.const -2)) (i64.const 0))
	// (assert_return (invoke "div_u" (i64.const -5) (i64.const -2)) (i64.const 0))
	// (assert_return (invoke "div_u" (i64.const 7) (i64.const 3)) (i64.const 2))

	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 7)
	vm.locals = append(vm.locals, 3)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(2))

	// (assert_return (invoke "div_u" (i64.const 11) (i64.const 5)) (i64.const 2))
	vm.pointInCode = 0
	vm.locals = []uint64{}
	vm.locals = append(vm.locals, 11)
	vm.locals = append(vm.locals, 5)

	vm.callStack[0].Ip = 0
	vm.currentFrame = 0
	vm.callStack[0].Locals = vm.locals

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(2))

	// (assert_return (invoke "div_u" (i64.const 17) (i64.const 7)) (i64.const 2))

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
