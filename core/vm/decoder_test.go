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
	expectedCodeSection := []byte {
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