package vm

import (
	"testing"
)

func Test_newVirtualMachine(t *testing.T) {
	vm := newVirtualMachine(generateTestWasm(), Storage{})
	// TODO: make this fail. Once the wasm is actually ran, it will fail!
	println(vm)
	vm.step()
	println(len(vm.vmMemory) == 0)
	vm.run()
	println(len(vm.vmMemory) == 0)
	// if we make it this far, ill count it as a success
}
func Test_parseCodeToOpcodes(t *testing.T) {
	answer := []int8{0, 97, 115, 109, 1, 0, 0, 0, 1, 0, 1, 96, 1, 115, 1, 115, 6, 3, 0, 1, 0, 2, 10, 0, 1}
	result := parseCodeToOpcodes(generateTestWasm())
	if result[0].opcode != int8(answer[0]) {
		t.Fail()
	}
	for i, v := range result[0].params {
		if v != int8(answer[i+1]) {
			t.Fail()
		}

	}
}

func generateTestWasm() []string {
	// https://en.wikipedia.org/wiki/WebAssembly code grabbed from
	return []string{
		"00 61 73 6D 01 00 00 00",
		"01 00 01 60 01 73 01 73 06",
		"03 00 01 00 02",
		"0A 00 01",
		"00 00",
		"20 00",
		"50",
		"04 7E",
		"42 01",
		"05",
		"20 00",
		"20 00",
		"42 01",
		"7D",
		"10 00",
		"7E",
		"0B",
		"0B 15 17"}
}

func Test_virtualMachineWithMyCode(t *testing.T) {
	wasm := []string{
		"7e"}

	vm := newVirtualMachine(wasm, Storage{})
	vm.run()
	// vm.do(operation{opcode: 0x7e})

	println("vm memory consists of")
	println(vm.outputStack())
	if vm.outputStack() != "0\n" {
		t.Fail()
	}
}
