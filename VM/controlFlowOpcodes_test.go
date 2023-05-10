package VM

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/stretchr/testify/assert"
)

// func Test_controlFlowOpcodes(t *testing.T) {
// 	moduleCodes := [][]byte{}
// 	wasmStrings := []string{}

// }
func Test_OpBlock(t *testing.T) {
	t.Parallel()
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f03828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000ab28080800001ac8080800001017f410028020441106b2200410036020c024041010d002000200028020c41136c36020c0b200028020c0b")

	_ = NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	module := *decode(wasmBytes)

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

	assert.Equal(t, expectedModuleCode, module.codeSection[0].body)
}

func Test_SingleBlock(t *testing.T) {
	t.Parallel()
	wasmBytes, _ := hex.DecodeString("0061736d0100000001060160027f7f0003020100070a010661646454776f00000a0d010b000240410a410f6a1a0b0b000a046e616d650203010000")
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	module := *decode(wasmBytes)

	expectedModuleCode := []byte{
		Op_block, Op_empty,
		Op_i32_const, 0xa,
		Op_i32_const, 0xf,
		Op_i32_add,
		Op_drop,
		Op_end,
		Op_end,
	}

	assert.Equal(t, expectedModuleCode, module.codeSection[0].body)
	vm.run()
}

func Test_MultiBlock(t *testing.T) {
	t.Parallel()
	wasmBytes, _ := hex.DecodeString("0061736d0100000001060160027f7f0003020100070a010661646454776f00000a10010e0002400240410a410f6a1a0b0b0b000a046e616d650203010000")
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

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
		Op_i32_const, 0x2,
		Op_i32_const, 0x3,
		Op_i32_mul,
		Op_drop,
		Op_end,

		Op_nop,
		Op_end,
	}

	bytes, cs := parseBytes(code)

	vm.vmCode = bytes
	vm.controlBlockStack = cs

	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack

	vm.run()
	assert.Equal(t, len(vm.vmStack), 0)
}

func Test_Br(t *testing.T) {
	t.Parallel()
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f03828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000ab28080800001ac8080800001017f410028020441106b2200410036020c024041010d002000200028020c41146a36020c0b200028020c0b")
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

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
	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack

	vm.run()
	assert.Equal(t, vm.popFromStack(), uint64(0))
}

func Test_Br2(t *testing.T) {
	t.Parallel()
	wasmBytes, _ := hex.DecodeString("0061736d010000000186808080000160017f017f0382808080000100048480808000017000000583808080000100010681808080000007988080800002066d656d6f727902000b5f5a376d616b654164646900000ab98080800001b38080800001017f410028020441106b2201200036020c2001410a360208024041000d002001200128020841146a3602080b20012802080b")
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

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
	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack
	vm.run()
	fmt.Printf("vmStack: %v\n", vm.vmStack)
	assert.Equal(t, vm.popFromStack(), uint64(30))
}

func Test_Loop(t *testing.T) {
	t.Parallel()
	// int testFunction() {
	// 	int sum = 0;
	// 	for(int i = 0; i < 10; ++i) {
	// 	  sum += i;
	// 	}
	// 	return sum;
	// }
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f03828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000ad48080800001ce8080800001017f410028020441106b2200410036020c2000410036020802400340200028020841094a0d012000200028020c20002802086a36020c2000200028020841016a3602080c000b0b200028020c0b")
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

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
	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack

	vm.run()
	fmt.Printf("vmStack: %v\n", vm.vmStack)
	assert.Equal(t, vm.popFromStack(), uint64(45))
}

func Test_If(t *testing.T) {
	t.Parallel()
	wasmBytes, _ := hex.DecodeString("0061736d010000000184808080000160000003828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000abd8080800001b78080800001017f410028020441106b220041e40036020c02404100450d002000200028020c410a6a36020c0f0b2000200028020c410f6a36020c0b")
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

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
	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack

	vm.run()
	res := LE.Uint32(vm.vmMemory[12 : 12+4])
	assert.Equal(t, uint64(115), uint64(res))
}

func Test_Return(t *testing.T) {
	t.Parallel()
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
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

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
	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack

	vm.run()
	res := LE.Uint32(vm.vmMemory[12 : 12+4])
	assert.Equal(t, uint64(110), uint64(res))
}

func Test_Call(t *testing.T) {
	t.Parallel()
	// (module
	// 	(func $fac (export "fac") (param f64) (result f64)
	// 	  local.get 0
	// 	  f64.const 1
	// 	  f64.lt
	// 	  if (result f64)
	// 		f64.const 1
	// 	  else
	// 		local.get 0
	// 		local.get 0
	// 		f64.const 1
	// 		f64.sub
	// 		call $fac
	// 		f64.mul
	// 	  end))

	wasmBytes, _ := hex.DecodeString("0061736d0100000001060160017c017c030201000707010366616300000a2e012c00200044000000000000f03f63047c44000000000000f03f052000200044000000000000f03fa11000a20b0b0012046e616d6501060100036661630203010000")
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

	expected := []byte{
		Op_get_local, 0x0,
		Op_f64_const, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f,
		Op_f64_lt,

		Op_if, Op_f64,
		Op_f64_const, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f,
		Op_else,
		Op_get_local, 0x0,
		Op_get_local, 0x0,
		Op_f64_const, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f,
		Op_f64_sub,
		Op_call, 0x0,
		Op_f64_mul,
		Op_end,

		Op_end,
	}

	module := *decode(wasmBytes)
	assert.Equal(t, expected, module.codeSection[0].body)
	vm.AddLocal(float64(5))
	vm.callStack[0].Locals = vm.locals

	callerAddr := common.BytesToAddress([]byte{0x1, 0x2, 0x3, 0x4})
	var gas = big.NewInt(100)
	contract := newContract(callerAddr, gas, module.codeSection[0].body, 100)
	spoofer := NewDBSpoofer()

	localCodeStored := CodeStored{
		CodeParams:  module.typeSection[0].params,
		CodeResults: module.typeSection[0].results,
		CodeBytes:   module.codeSection[0].body,
	}
	localCodeStoredHash, _ := localCodeStored.Hash()
	contract.CodeHashes = []string{hex.EncodeToString(localCodeStoredHash)}
	spoofer.AddSpoofedCode(hex.EncodeToString(localCodeStoredHash), localCodeStored)
	vm.config.CodeGetter = spoofer.GetCode

	vm.contract = *contract

	vm.callStack[vm.currentFrame].Code, vm.controlBlockStack = parseBytes(module.codeSection[0].body)
	vm.run()

	assert.Equal(t, math.Float64frombits(vm.popFromStack()), float64(120))

	vm = NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	vm.config.CodeGetter = spoofer.GetCode
	vm.contract = *contract
	vm.AddLocal(float64(8))
	vm.callStack[0].Locals = vm.locals
	vm.callStack[vm.currentFrame].Code, vm.controlBlockStack = parseBytes(module.codeSection[0].body)

	vm.run()

	assert.Equal(t, math.Float64frombits(vm.popFromStack()), float64(40320))

	vm = NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	vm.config.CodeGetter = spoofer.GetCode
	vm.contract = *contract
	vm.AddLocal(float64(12))
	vm.callStack[0].Locals = vm.locals
	vm.callStack[vm.currentFrame].Code, vm.controlBlockStack = parseBytes(module.codeSection[0].body)

	vm.run()
	assert.Equal(t, math.Float64frombits(vm.popFromStack()), float64(479001600))

	vm = NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	vm.config.CodeGetter = spoofer.GetCode
	vm.contract = *contract
	vm.AddLocal(float64(14))
	vm.callStack[0].Locals = vm.locals
	vm.callStack[vm.currentFrame].Code, vm.controlBlockStack = parseBytes(module.codeSection[0].body)

	vm.run()

	assert.Equal(t, math.Float64frombits(vm.popFromStack()), float64(87178291200))

	vm = NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	vm.config.CodeGetter = spoofer.GetCode
	vm.contract = *contract
	vm.AddLocal(float64(25))
	vm.callStack[0].Locals = vm.locals
	vm.callStack[vm.currentFrame].Code, vm.controlBlockStack = parseBytes(module.codeSection[0].body)

	vm.run()

	assert.Equal(t, math.Float64frombits(vm.popFromStack()), float64(15511210043330985984000000))

	//floating math only holds accurate to 27!
	vm = NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	vm.config.CodeGetter = spoofer.GetCode
	vm.contract = *contract
	vm.AddLocal(float64(27))
	vm.callStack[0].Locals = vm.locals
	vm.callStack[vm.currentFrame].Code, vm.controlBlockStack = parseBytes(module.codeSection[0].body)

	vm.run()
	assert.Equal(t, math.Float64frombits(vm.popFromStack()), float64(10888869450418352160768000000))
}

func Test_FuncFact(t *testing.T) {
	t.Parallel()
	// double fact() {
	// 	int i = 4;
	// 	long long n = 1;
	// 	for (;i > 0; i--) {
	// 	  n *= i;
	// 	}
	// 	return (double)n;
	//   }
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017c0382808080000100048480808000017000000583808080000100010681808080000007958080800002066d656d6f72790200085f5a34666163747600000ad58080800001cf8080800001017f410028020441106b2200410436020c2000420137030002400340200028020c4101480d0120002000290300200034020c7e3703002000200028020c417f6a36020c0c000b0b2000290300b90b")
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

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
		0x2, 0xc, 0xc, 0x0, 0xb, 0xb, 0x20, 0x0, 0x29, 0x3, 0x0, 0xb9, 0xb,
	}

	vm.vmCode, vm.controlBlockStack = parseBytes(expected)
	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack
	vm.run()
	fmt.Printf("vm.vmStack: %v\n", vm.vmStack)
	res := vm.popFromStack()
	assert.Equal(t, res, uint64(24))
}

func Test_blockDeep(t *testing.T) {
	t.Parallel()
	// https://github.com/WebAssembly/testsuite//blob/main/block.wast#L40
	wasmBytes, _ := hex.DecodeString("0061736d010000000108026000006000017f0303020001070801046465657000010a7e0202000b7900027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f027f10004196010b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0016046e616d65010801000564756d6d7902050200000100")
	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

	expected := []byte{
		0x2, 0x7f,
		0x2, 0x7f, 0x2,
		0x7f, 0x2, 0x7f,
		0x2, 0x7f, 0x2,
		0x7f, 0x2, 0x7f,
		0x2, 0x7f, 0x2,
		0x7f, 0x2, 0x7f, 0x2,
		0x7f, 0x2, 0x7f, 0x2,
		0x7f, 0x2, 0x7f,
		0x2, 0x7f, 0x2,
		0x7f, 0x2, 0x7f,
		0x2, 0x7f, 0x2, 0x7f,
		0x2, 0x7f, 0x2, 0x7f,
		0x2, 0x7f, 0x2, 0x7f, 0x2,
		0x7f, 0x2, 0x7f, 0x2, 0x7f,
		0x2, 0x7f, 0x2, 0x7f, 0x2, 0x7f, 0x2,
		0x7f, 0x2, 0x7f, 0x2, 0x7f, 0x2, 0x7f, 0x2, 0x7f, 0x2, 0x7f, 0x2, 0x7f, 0x2,
		0x7f, 0x2, 0x7f, 0x10, 0x0, 0x41, 0x96, 0x1, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb,
		0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb,
		0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb, 0xb}

	vm.vmCode, vm.controlBlockStack = parseBytes(expected)
	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack

	module := *decode(wasmBytes)
	vm.run()
	assert.Equal(t, expected, module.codeSection[1].body)
	assert.Equal(t, vm.popFromStack(), uint64(150))
}

func Test_blockEmpty(t *testing.T) {
	t.Parallel()
	wasmBytes, _ := hex.DecodeString("0061736d010000000104016000000302010007090105656d70747900000a0a01080002400b02400b0b000a046e616d650203010000")

	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

	module := *decode(wasmBytes)
	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[0].body)
	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack
	err := vm.run()
	fmt.Printf("err: %v\n", err)
}

func Test_blockNested(t *testing.T) {
	t.Parallel()
	wasmBytes, _ := hex.DecodeString("0061736d010000000108026000006000017f0303020001070a01066e657374656400010a1a0202000b1500027f0240100002400b010b027f100041090b0b0b0016046e616d65010801000564756d6d7902050200000100")

	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

	module := *decode(wasmBytes)
	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[1].body)
	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack
	err := vm.run()
	fmt.Printf("err: %v\n", err)
}

func Test_blockAsLoop(t *testing.T) {
	t.Parallel()
	wasmBytes, _ := hex.DecodeString("0061736d010000000108026000006000017f03030200010711010d61732d6c6f6f702d666972737400010a130202000b0e00037f027f41010b100010000b0b0016046e616d65010801000564756d6d7902050200000100")

	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	vm.contract = *testContract

	module := *decode(wasmBytes)
	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[1].body)
	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack
	err := vm.run()
	fmt.Printf("err: %v\n", err)
}

func Test_LoopDeep(t *testing.T) {
	t.Parallel()
	// https://github.com/WebAssembly/testsuite//blob/main/loop.wast#L37
	wasmBytes, _ := hex.DecodeString("0061736d010000000108026000006000017f0303020001070801046465657000010a84010202000b7f00037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f037f027f10004196010b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0016046e616d65010801000564756d6d7902050200000100")

	vm := NewVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	vm.contract = *testContract

	module := *decode(wasmBytes)
	vm.vmCode, vm.controlBlockStack = parseBytes(module.codeSection[1].body)
	vm.callStack[0].Code, vm.callStack[0].CtrlStack = vm.vmCode, vm.controlBlockStack
	err := vm.run()
	res := vm.popFromStack()
	assert.Equal(t, res, uint64(0x96))
	fmt.Printf("err: %v\n", err)
}
