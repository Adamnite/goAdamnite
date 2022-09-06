package vm

import (
	"testing"
)

// func Test_newVirtualMachine(t *testing.T) {
// 	// TODO: make this fail. Once the wasm is actually ran, it will fail!
// 	println(doesStackMatchExpected(generateTestWasm(), []uint64{2}, false, []uint64{2}))
// 	println(doesStackMatchExpected(generateTestWasm(), []uint64{1}, false, []uint64{0}))
// 	// println(vm.outputStack())
// 	// if we make it this far, ill count it as a success
// }

// func Test_parseCodeToOpcodes(t *testing.T) {
// 	opcodeAns := []uint8{0x00, 0x01, 0x01, 0x01, 0x01, 0x03, 0x0A, 0x00, 0x20, 0x50, 0x04, 0x42, 0x05, 0x20, 0x20, 0x42, 0x7D, 0x10, 0x7E, 0x0B, 0x0B}
// 	answer := [][]uint8{{0x61, 0x73, 0x6D},
// 		{0x00, 0x00, 0x00},
// 		{0x00},
// 		{0x60},
// 		{0x73, 0x01, 0x73, 0x06},
// 		{0x00, 0x01, 0x00, 0x02},
// 		{0x00, 0x01},
// 		{0x00},
// 		{0x00},
// 		{},
// 		{0x7E},
// 		{0x01},
// 		{},
// 		{0x00},
// 		{0x00},
// 		{0x01},
// 		{},
// 		{0x00},
// 		{},
// 		{},
// 		{0x15, 0x17}}
// 	result := parseCodeToOpcodes(generateTestWasm())
// 	for i, v := range result {
// 		// print("index: ")
// 		// print(i)
// 		// print("      value: ")
// 		// println(v.opcode)
// 		if v.opcode != opcodeAns[i] {
// 			//check the opcodes are correct
// 			t.Fail()
// 		}
// 		for j, m := range v.params {
// 			if answer[i][j] != m {
// 				t.Fail()
// 			}
// 		}
// 	}
// }//if commented out, generateTestWasm has most likely been reformatted

// https://webassembly.github.io/spec/core/appendix/index-instructions.html
func generateTestWasm() []string {
	// https://en.wikipedia.org/wiki/WebAssembly code grabbed from
	return []string{
		// "00 61 73 6D", //WASM binary magic!
		// "01 00 00 00", //wasm binary version
		// space hear to clearly remove header
		"01 00",          //section code, size (guess) 0
		"01 60",          //"num types", func
		"01 73 01 73 06", //nop, unclear
		"03 00 01 00 02", //loop
		"0A 00 01",       //0A is reserved, unclear why it is here.
		"00 00",          //unreachable
		"20 00",          //get local 00
		"50",             //i64 eqz
		"04 7E",          //if, with the popped value type i64. runs if not 0
		"42 01",          //i64 const, val 1
		"05",             //else
		"20 00",          //get local
		"20 00",          //get local
		"42 01",          //i64 const 01
		"7D",             //i64_sub
		"10 00",          //op_call block 00
		"7E",             //i64, or i64.mul
		"0B",             //end
		"0B 15 17"}       //end 15, 17
}

func doesStackMatchExpected(wasm []OperationCommon, ansStack []uint64, debug bool, locals []uint64) bool {
	vm := newVirtualMachine(wasm, Storage{})

	vm.debugStack = debug
	vm.locals = locals
	allClear := true
	vm.run()
	for i, x := range ansStack {
		allClear = allClear && (vm.vmStack[i] == x)
	}
	return allClear
}

func Test_virtualMachineWithBasicIfCaseCode(t *testing.T) {
	// wasm := []string{
	// 	"42 05", //i64.const 0x05
	// 	"04 7E", //make an if statement that will run (the top value is 0x05)
	// 	"42 F0", //there should be a 0xF0 on the stack due to this
	// 	"05",    //else case that shouldnt run
	// 	"42 FF", //shouldnt see any FF
	// 	"0B",    //put an end for that if statement.
	// 	"42 00", //push 0x00 to the stack, lets test the else side
	// 	"04 7E", //test the top value is not equal to 00(will fail)
	// 	"42 FF", //if we see 0xFF in the stack at all, assume a failure.
	// 	"05",    //else statement to if on line 5
	// 	"42 F0", //hope to see the stack total being F0 F0 at the end
	// 	"0B",    //end to this if statement
	// }
	wasm := parseString("42 00 42 01 7c")
	// ansStack := []uint64{0xF0, 0xF0}
	println(len(wasm))
	vm := newVirtualMachine(wasm, Storage{})
	vm.debugStack = true
	// vm.step()
	// vm.step()
	vm.run()
	// if !doesStackMatchExpected(wasm, ansStack, false, []uint64{}) {
	// 	t.Fail()
	// }
}
