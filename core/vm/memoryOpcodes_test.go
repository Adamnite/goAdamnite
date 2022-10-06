package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


func Test_i32Store(t *testing.T) {
	wasm := parseString("41 0c 41 01 41 02 36")

	vm := newVirtualMachine(wasm, Storage{}, VMConfig{})
	vm.debugStack = true
	vm.run()
	r := LE.Uint32(vm.vmMemory[3 : 7])
	assert.Equal(t, r, uint32(12))
}


func Test_Opi32load(t *testing.T) {
	wasm := parseString("41 0c 41 01 41 02 36 41 0c 41 01 41 02 28")

	vm := newVirtualMachine(wasm, Storage{}, VMConfig{})
	vm.debugStack = true
	vm.run()
	res := vm.popFromStack()
	assert.Equal(t, uint32(res), uint32(12))
}

func Test_i64Store(t *testing.T) {
	wasm := parseString("41 0c 41 01 41 02 37")

	vm := newVirtualMachine(wasm, Storage{}, VMConfig{})
	vm.debugStack = true
	vm.run()
	r := LE.Uint64(vm.vmMemory[3 : 11])
	assert.Equal(t, r, uint64(12))
}


func Test_Opi64load(t *testing.T) {
	wasm := parseString("41 0c 41 01 41 02 37 41 0c 41 01 41 02 29")

	vm := newVirtualMachine(wasm, Storage{}, VMConfig{})
	vm.debugStack = true
	vm.run()
	res := vm.popFromStack()
	assert.Equal(t, uint64(res), uint64(12))
} 