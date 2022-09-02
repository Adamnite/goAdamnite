package vm

import (
	"testing"
)

func Test_newVirtualMachine(t *testing.T) {
	vm := newVirtualMachine(generateTestWasm(), Storage{})
	// TODO: make this fail. Once the wasm is actually ran, it will fail!
	vm.step()
	println(len(vm.vmMemory) == 0)
	vm.run()
	println(len(vm.vmStack) != 0)
	println(len(vm.vmMemory) == 0)
	// if we make it this far, ill count it as a success
}
func Test_parseCodeToOpcodes(t *testing.T) {
	opcodeAns := []int8{0x00, 0x01, 0x01, 0x01, 0x01, 0x03, 0x0A, 0x00, 0x20, 0x50, 0x04, 0x42, 0x05, 0x20, 0x20, 0x42, 0x7D, 0x10, 0x7E, 0x0B, 0x0B}
	answer := [][]int8{{0x61, 0x73, 0x6D},
		{0x00, 0x00, 0x00},
		{0x00},
		{0x60},
		{0x73, 0x01, 0x73, 0x06},
		{0x00, 0x01, 0x00, 0x02},
		{0x00, 0x01},
		{0x00},
		{0x00},
		{},
		{0x7E},
		{0x01},
		{},
		{0x00},
		{0x00},
		{0x01},
		{},
		{0x00},
		{},
		{},
		{0x15, 0x17}}
	result := parseCodeToOpcodes(generateTestWasm())
	for i, v := range result {
		// print("index: ")
		// print(i)
		// print("      value: ")
		// println(v.opcode)
		if v.opcode != opcodeAns[i] {
			//check the opcodes are correct
			t.Fail()
		}
		for j, m := range v.params {
			if answer[i][j] != m {
				t.Fail()
			}
		}
	}
}

// https://webassembly.github.io/spec/core/appendix/index-instructions.html
func generateTestWasm() []string {
	// https://en.wikipedia.org/wiki/WebAssembly code grabbed from
	return []string{
		"00 61 73 6D",    //WASM binary magic!
		"01 00 00 00",    //wasm binary version
		"01 00",          //section code, size (guess) 0
		"01 60",          //"num types", func
		"01 73 01 73 06", //nop, unclear
		"03 00 01 00 02", //loop
		"0A 00 01",       //0A is reserved, unclear why it is here.
		"00 00",          //unreachable
		"20 00",          //block 00
		"50",             //i64 eqz
		"04 7E",          //if, either i64, or i64_mul
		"42 01",          //i64 const, val 1
		"05",             //else
		"20 00",          //get local
		"20 00",          //get local
		"42 01",          //i64 const 01
		"7D",             //i64_sub or op_f32
		"10 00",          //op_call block 00
		"7E",             //i64, or i64.mul
		"0B",             //end
		"0B 15 17"}       //end 15, 17
}

func Test_virtualMachineWithMyCode(t *testing.T) {
	wasm := []string{
		"42 05",
		"50"}

	vm := newVirtualMachine(wasm, Storage{})
	vm.run()

	// println("vm stack consists of")
	// println(vm.outputStack())
	// println(vm.outputMemory())
	if vm.outputStack() != "0\n" {
		t.Fail()
	}
}
