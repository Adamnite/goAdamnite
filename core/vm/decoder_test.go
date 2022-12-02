package vm

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecoderOrdinary(t *testing.T) {
	bytes := "0061736d0100000001070160027f7f017f03020100070a010661646454776f00000a09010700200020016a0b000a046e616d650203010000"

	ansBytes, _ := hex.DecodeString(bytes)

	res := decode(ansBytes)
	expectedCodeSection := []byte{
		Op_get_local,
		0x0,
		Op_get_local,
		0x01,
		Op_i32_add,
		Op_end,
	}

	assert.Equal(t, expectedCodeSection, res.codeSection[0].body)
}

func TestDecodeWithIf(t *testing.T) {
	bytes := "0061736d010000000186808080000160017f017f0382808080000100048480808000017000000583808080000100010681808080000007948080800002066d656d6f72790200076d616b6541646400000ab98080800001b38080800001017f410028020441106b2201200036020820014114360204024041010d002001200128020441016a3602040b200128020c0b"
	ansBytes, _ := hex.DecodeString(bytes)

	expectedCodeSection := []byte{
		Op_i32_const, 0x0,
		Op_i32_load, 0x2, 0x4,
		Op_i32_const, 0x10,
		Op_i32_sub,
		Op_tee_local, 0x1,
		Op_get_local, 0x0,
		Op_i32_store, 0x2, 0x8,
		Op_get_local, 0x1,
		Op_i32_const, 0x14,
		Op_i32_store, 0x2, 0x4,
		Op_block, 0x40,

		Op_i32_const, 0x1,
		Op_br_if,
		Op_unreachable,
		Op_get_local, 0x1,
		Op_get_local, 0x1,
		Op_i32_load, 0x2, 0x4,

		Op_i32_const, 0x1,
		Op_i32_add,
		Op_i32_store, 0x2, 0x4,
		Op_end,

		Op_get_local, 0x1,
		Op_i32_load, 0x2, 0xc,
		Op_end}

	res := decode(ansBytes)
	assert.Equal(t, expectedCodeSection, res.codeSection[0].body)
}

func Test_decode2(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f03828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000ab28080800001ac8080800001017f410028020441106b2200410036020c024041010d002000200028020c41136c36020c0b200028020c0b")

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

func Test_i32StoreDecode(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f0382808080000100048480808000017000000583808080000100010681808080000007918080800002066d656d6f72790200046d61696e00000aa280808000019c8080800001017f410028020441106b2200410036020c2000410a360208410a0b")

	vm := newVirtualMachine(wasmBytes, []uint64{}, nil, 1000)
	
	expectedModuleCode := []byte{
		Op_i32_const, 0x0,
		Op_i32_load, 0x2, 0x4,
		Op_i32_const, 0x10,
		Op_i32_sub,
		Op_tee_local, 0x0,
		Op_i32_const, 0x0,
		Op_i32_store, 0x2, 0xc,
		Op_get_local, 0x0,
		Op_i32_const, 0xa,
		Op_i32_store, 0x2, 0x8,
		Op_i32_const, 0xa,
		Op_end,
	}
	assert.Equal(t, expectedModuleCode, vm.module.codeSection[0].body)
}

func Test_i32Store3Decode(t *testing.T) {
	wasmBytes, _ := hex.DecodeString("0061736d01000000018580808000016000017f03828080800001000484808080000170000005838080800001000106818080800000079e8080800002066d656d6f72790200115f5a31327465737446756e6374696f6e7600000a978080800001918080800000410028020441106b410436020800000b")

	vm := newVirtualMachine(wasmBytes, []uint64{}, nil, 1000)

	expectedModuleCode := []byte{
		Op_i32_const, 0x0,
		Op_i32_load, 0x2, 0x4,
		Op_i32_const, 0x10,
		Op_i32_sub,
		Op_i32_const, 0x4,
		Op_i32_store, 0x2, 0x8,
		0x0, 0x0,
		Op_end,
	}

	assert.Equal(t, expectedModuleCode, vm.module.codeSection[0].body)
}
