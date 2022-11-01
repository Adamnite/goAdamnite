package vm

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OpBlock(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f03828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000ab28080800001ac8080800001017f410028020441106b2200410036020c024041010d002000200028020c41136c36020c0b200028020c0b")

	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})
	vm.debugStack = true
	expectedModuleCode := []byte{
		Op_i32_const, 0x0,
		Op_i32_load, 0x2, 0x4,
		Op_i32_const, 0x10,
		Op_i32_sub,
		Op_tee_local, 0x0,
		Op_i32_const, 0x0,
		Op_i32_store, 0x2, 0xc,

		Op_block, 0x40,
		Op_i32_const, 0x1,
		Op_br_if, 0x0,
		Op_get_local, 0x0,
		Op_get_local, 0x0,
		Op_i32_load, 0x2, 0xc,
		Op_i32_const, 0x13,
		Op_i32_mul,
		Op_i32_store, 0x2, 0xc,

		Op_end,
		Op_get_local, 0x0,
		Op_i32_load, 0x2, 0xc,
		Op_end,
	}

	assert.Equal(t, expectedModuleCode, vm.module.codeSection[0].body)
}

func Test_SingleBlock(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d0100000001060160027f7f0003020100070a010661646454776f00000a0d010b000240410a410f6a1a0b0b000a046e616d650203010000")
	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})
	vm.debugStack = true

	expectedModuleCode := []byte{
		Op_block, Op_empty,
		Op_i32_const, 0xa,
		Op_i32_const, 0xf,
		Op_i32_add,
		Op_drop,
		Op_end,
		Op_end,
	}

	assert.Equal(t, expectedModuleCode, vm.module.codeSection[0].body)
	vm.run()
}

func Test_MultiBlock(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d0100000001060160027f7f0003020100070a010661646454776f00000a10010e0002400240410a410f6a1a0b0b0b000a046e616d650203010000")
	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})
	vm.debugStack = true

	code := []byte{
		Op_block, Op_empty,
			Op_i32_const, 0xa,
			Op_i32_const, 0xf,
			Op_i32_add,
			Op_drop,

			Op_block, Op_empty,
				Op_i32_const, 0x2,
				Op_i32_const, 0x3,
				Op_i32_mul,
				Op_drop,
			Op_end,

			Op_nop,

		Op_end,
		Op_end,
	}

	bytes, cs := parseBytes(code)

	vm.vmCode = bytes
	vm.controlBlockStack = cs
	vm.run()
	assert.Equal(t, len(vm.vmStack), 0)
}

func Test_Br(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f03828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000ab28080800001ac8080800001017f410028020441106b2200410036020c024041010d002000200028020c41146a36020c0b200028020c0b")
	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})
	vm.debugStack = true

	code := []byte{
		Op_i32_const, 0x0,
		Op_i32_const, 0x10,
		Op_i32_load, 0x2, 0x4,
		Op_i32_sub,
		Op_tee_local, 0x0,
		Op_i32_const, 0x0,
		Op_i32_store, 0x2, 0xc,

		Op_block, Op_empty,
		Op_i32_const, 0x1,
		Op_br_if, 0x0,
		Op_get_local, 0x0,
		Op_get_local, 0x0,
		Op_i32_load, 0x2, 0xc,
		Op_i32_const, 0x14,
		Op_i32_add,

		Op_i32_store, 0x2, 0xc,
		Op_end,
		Op_get_local, 0x0,
		Op_i32_load, 0x2, 0xc,
		Op_end,
	}

	vm.vmCode, vm.controlBlockStack = parseBytes(code)
	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0))
}

func Test_Br2(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d010000000186808080000160017f017f0382808080000100048480808000017000000583808080000100010681808080000007988080800002066d656d6f727902000b5f5a376d616b654164646900000ab98080800001b38080800001017f410028020441106b2201200036020c2001410a360208024041000d002001200128020841146a3602080b20012802080b")	
	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})
	vm.debugStack = true

	code := []byte{
		Op_i32_const, 0x0,
		Op_i32_const, 0x10,
		Op_i32_load, 0x2, 0x4, 
		Op_i32_sub, 
		Op_tee_local, 0x1, 
		
		Op_get_local, 0x0, 
		Op_i32_store, 0x2, 0xc, 
		
		Op_get_local, 0x1, 
		Op_i32_const, 0xa, 
		Op_i32_store, 0x2, 0x8, 
		
		Op_block, Op_empty, 
			Op_i32_const, 0x0, 
			Op_br_if, 0x0, 
			Op_get_local, 0x1,
			Op_get_local, 0x1, 
			Op_i32_load, 0x2, 0x8,
			Op_i32_const, 0x14, 
			Op_i32_add, 
			Op_i32_store, 0x2, 0x8, 
		Op_end,
		
		Op_get_local, 0x1, 
		Op_i32_load, 0x2, 0x8, 
		
		Op_end,
	}

	vm.vmCode, vm.controlBlockStack = parseBytes(code)
	vm.run()
	fmt.Printf("vmStack: %v\n", vm.vmStack)
	assert.Equal(t, vm.popFromStack(), uint64(30))
}


func Test_Loop(t *testing.T) {
	// int testFunction() {
	// 	int sum = 0;
	// 	for(int i = 0; i < 10; ++i) {
	// 	  sum += i;
	// 	}
	// 	return sum;
	// }
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f03828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000ad48080800001ce8080800001017f410028020441106b2200410036020c2000410036020802400340200028020841094a0d012000200028020c20002802086a36020c2000200028020841016a3602080c000b0b200028020c0b")	
	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})
	vm.debugStack = true

	code := []byte{
		0x41, 0x0, 
		0x41, 0x10, 
		0x28, 0x2, 0x4, 
		0x6b, 0x22, 0x0, 0x41, 0x0, 
		0x36, 0x2, 0xc, 
		0x20, 0x0, 0x41, 0x0, 
		0x36, 0x2, 0x8, 
		0x2, 0x40, 
		
		0x3, 0x40, 
		0x20, 0x0, 0x28, 
		0x2, 0x8, 0x41, 
		0x9, 0x4a, 
		0xd, 0x1, 0x20, 
		0x0, 0x20, 0x0, 
		0x28, 0x2, 0xc, 
		0x20, 0x0, 0x28, 
		0x2, 0x8, 0x6a, 
		0x36, 0x2, 0xc, 
		0x20, 0x0, 0x20, 0x0, 
		0x28, 0x2, 0x8, 
		0x41, 0x1, 0x6a, 
		0x36, 0x2, 0x8, 0xc, 
		0x0, 0xb, 0xb, 
		0x20, 0x0, 
		0x28, 0x2, 0xc, 0xb,
	}

	vm.vmCode, vm.controlBlockStack = parseBytes(code)
	vm.run()
	fmt.Printf("vmStack: %v\n", vm.vmStack)
	assert.Equal(t, vm.popFromStack(), uint64(45))
}

func Test_If(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d010000000184808080000160000003828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000abd8080800001b78080800001017f410028020441106b220041e40036020c02404100450d002000200028020c410a6a36020c0f0b2000200028020c410f6a36020c0b")	
	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})
	
	code := []byte{
		Op_i32_const, 0x0, 
		Op_i32_const, 0x10, 
		Op_i32_load, 0x2, 0x4, 
		Op_i32_sub, 
		
		Op_tee_local, 0x0, 
		Op_i32_const, 0xe4, 0x0, 
		Op_i32_store, 0x2, 0xc, 
		Op_block, 0x40, 
			Op_i32_const, 0x0, 
			Op_i32_eqz, 
			Op_br_if, 
			0x0, 
			Op_get_local, 0x0, 
			Op_get_local, 0x0, 
			Op_i32_load, 0x2, 0xc, 
			Op_i32_const, 0xa, 
			Op_i32_add, 
			Op_i32_store, 0x2, 0xc, 
			Op_return, 
		Op_end, 
			Op_get_local, 0x0, 
			Op_get_local, 0x0, 
			Op_i32_load, 0x2, 0xc, 
			Op_i32_const, 0xf, 
			Op_i32_add, 
			Op_i32_store, 0x2, 0xc, 
			
		Op_end,
	}

	vm.vmCode, vm.controlBlockStack = parseBytes(code)
	vm.run()
	res := LE.Uint32(vm.vmMemory[12 : 12 + 4])
	assert.Equal(t, uint64(115), uint64(res))
}

func Test_Return(t *testing.T) {
	// int testFunction() {
	// 	int sum = 100;
	// 	if (sum%2 != 0) {
	// 	  sum += 15;
	// 	} else {
	// 	  sum+= 10;
	// 	}
	// 	return sum;
	//   }
	wasmBytes, _ := hex.DecodeString("0061736d010000000184808080000160000003828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000abd8080800001b78080800001017f410028020441106b220041e40036020c02404100450d002000200028020c410a6a36020c0f0b2000200028020c410f6a36020c0b")	
	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})
	
	expected := []byte{
		Op_i32_const, 0x0, 
		Op_i32_const, 0x10, 
		Op_i32_load, 0x2, 0x4, 
		Op_i32_sub, 
		
		Op_tee_local, 0x0, 
		Op_i32_const, 0xe4, 0x0, 
		Op_i32_store, 0x2, 0xc, 
		Op_block, 0x40, 
			Op_i32_const, 0x1, 
			Op_i32_eqz, 
			Op_br_if, 
			0x0, 
			Op_get_local, 0x0, 
			Op_get_local, 0x0, 
			Op_i32_load, 0x2, 0xc, 
			Op_i32_const, 0xa, 
			Op_i32_add, 
			Op_i32_store, 0x2, 0xc, 
			Op_return, 
		Op_end, 
			Op_get_local, 0x0, 
			Op_get_local, 0x0, 
			Op_i32_load, 0x2, 0xc, 
			Op_i32_const, 0xf, 
			Op_i32_add, 
			Op_i32_store, 0x2, 0xc, 
			
		Op_end,
	}

	vm.vmCode, vm.controlBlockStack = parseBytes(expected)

	vm.run()
	res := LE.Uint32(vm.vmMemory[12 : 12 + 4])
	assert.Equal(t, uint64(110), uint64(res))
}


// func Test_Call(t *testing.T) {
// 	// (module
// 	// 	(func $fac (export "fac") (param f64) (result f64)
// 	// 	  local.get 0
// 	// 	  f64.const 1
// 	// 	  f64.lt
// 	// 	  if (result f64)
// 	// 		f64.const 1
// 	// 	  else
// 	// 		local.get 0
// 	// 		local.get 0
// 	// 		f64.const 1
// 	// 		f64.sub
// 	// 		call $fac
// 	// 		f64.mul
// 	// 	  end))
	  
// 	wasmBytes, _ := hex.DecodeString("0061736d0100000001060160017c017c030201000707010366616300000a2e012c00200044000000000000f03f63047c44000000000000f03f052000200044000000000000f03fa11000a20b0b0012046e616d6501060100036661630203010000")
// 	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})

// 	expected := []byte{
// 		Op_get_local, 0x0, 
// 		Op_f64_const, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f, 
// 		Op_f64_lt, 
		
// 		Op_if, Op_f64, 
// 			Op_f64_const, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f, 
// 		Op_else, 
// 			Op_get_local, 0x0, 
// 			Op_get_local, 0x0, 
// 			Op_f64_const, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f, 
// 			Op_f64_sub, 
// 			Op_call, 0x0, 
// 			Op_f64_mul, 
// 		Op_end, 
		
// 		Op_end,
// 	}
	
// 	assert.Equal(t, expected, vm.module.codeSection[0].body)
// 	vm.locals = append(vm.locals, 5)
// 	vm.run()
// 	fmt.Printf("vm.vmStack: %v\n", vm.vmStack)
// }

func Test_FuncFact(t *testing.T) {
	// double fact() {
	// 	int i = 4;
	// 	long long n = 1;
	// 	for (;i > 0; i--) {
	// 	  n *= i;
	// 	}
	// 	return (double)n;
	//   }
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017c0382808080000100048480808000017000000583808080000100010681808080000007958080800002066d656d6f72790200085f5a34666163747600000ad58080800001cf8080800001017f410028020441106b2200410436020c2000420137030002400340200028020c4101480d0120002000290300200034020c7e3703002000200028020c417f6a36020c0c000b0b2000290300b90b")
	vm := newVirtualMachine(wasmBytes, Storage{}, VMConfig{})

	expected := []byte{
		0x41, 0x0, 
		0x41, 0x10,
		0x28, 0x2, 0x4,  
		0x6b, 0x22, 0x0, 0x41, 
		0x4, 0x36, 0x2, 0xc, 
		0x20, 0x0, 0x42, 0x1, 0x37, 
		0x3, 0x0, 0x2, 0x40, 
		0x3, 0x40, 0x20, 0x0, 0x28, 
		0x2, 0xc, 0x41, 0x1, 0x48, 0xd, 0x1, 0x20, 0x0, 
		0x20, 0x0, 0x29, 0x3, 0x0, 0x20, 0x0, 0x34, 
		0x2, 0xc, 0x7e, 0x37, 0x3, 0x0, 0x20, 0x0, 
		0x20, 0x0, 0x28, 0x2, 0xc, 0x41, 0x7f, 0x6a, 0x36, 
		0x2, 0xc, 0xc, 0x0, 0xb, 0xb, 0x20, 0x0, 0x29, 0x3, 0x0, 0xb9, 0xb}

	vm.vmCode, vm.controlBlockStack = parseBytes(expected)
	vm.run()
	fmt.Printf("vm.vmStack: %v\n", vm.vmStack)
	res := vm.popFromStack()
	assert.Equal(t, res, uint64(24))
}

