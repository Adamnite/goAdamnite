package vm

import (
	"encoding/hex"
	"reflect"
	"strings"
)

const wasmMagic = "0061736D01000000"

var wasmMagicBytes = []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}

func parseString(opString string) []OperationCommon {
	//sanitize the input of all possible special characters. Mostly used for tests
	s := strings.ReplaceAll(opString, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	if len(s) <= len(wasmMagic) {
		s = wasmMagic + s
	}
	if s[0:16] != wasmMagic {
		s = wasmMagic + s //if it doesn't have the wasm magic, add it
	}

	ansBytes, err := hex.DecodeString(s)
	if err != nil {
		println("error parsing string to bytes")
		println(err.Error())
		panic(err)
	}
	return parseBytes(ansBytes)
}

func parseBytes(bytes []byte) []OperationCommon {
	ansOps := []OperationCommon{}
	pointInBytes := 0
	if bytes[1] == 0x61 { //TODO: have this check the whole WASM magic number
		pointInBytes += 8
	}
	// println(hex.EncodeToString(bytes))
	for pointInBytes < len(bytes) {
		switch bytes[pointInBytes] {

		case Op_i32_const:
			ansOps = append(ansOps, i32Const{int32(bytes[pointInBytes+1])})
			pointInBytes += 2
		case Op_i32_sub:
			ansOps = append(ansOps, i32Sub{})
			pointInBytes += 1
		case Op_i32_add:
			ansOps = append(ansOps, i32Add{})
			pointInBytes += 1
		case Op_i32_div_s:
			ansOps = append(ansOps, i32Divs{})
			pointInBytes += 1
		case Op_i32_clz:
			ansOps = append(ansOps, i32Clz{})
			pointInBytes += 1
		case Op_i32_ctz:
			ansOps = append(ansOps, i32Ctz{})
			pointInBytes += 1

		case Op_i32_popcnt:
			ansOps = append(ansOps, i32PopCnt{})
			pointInBytes += 1
		case Op_i32_mul:
			ansOps = append(ansOps, i32Mul{})
			pointInBytes += 1

		case Op_i32_rem_s:
			ansOps = append(ansOps, i32Rems{})
			pointInBytes += 1

		case Op_i32_rem_u:
			ansOps = append(ansOps, i32Remu{})
			pointInBytes += 1

		case Op_i32_and:
			ansOps = append(ansOps, i32And{})
			pointInBytes += 1

		case Op_i32_or:
			ansOps = append(ansOps, i32Or{})
			pointInBytes += 1
		case Op_i32_xor:
			ansOps = append(ansOps, i32Xor{})
			pointInBytes += 1

		case Op_i32_shl:
			ansOps = append(ansOps, i32Shl{})
			pointInBytes += 1

		case Op_i32_shr_s:
			ansOps = append(ansOps, i32Shrs{})
			pointInBytes += 1

		case Op_i32_shr_u:
			ansOps = append(ansOps, i32Shru{})
			pointInBytes += 1

		case Op_i32_rotl:
			ansOps = append(ansOps, i32Rotl{})
			pointInBytes += 1

		case Op_i32_rotr:
			ansOps = append(ansOps, i32Rotr{})
			pointInBytes += 1

		case Op_i32_div_u:
			ansOps = append(ansOps, i32Divu{})
			pointInBytes += 1 

		case Op_i32_eqz:
			ansOps = append(ansOps, i32Eqz{})
			pointInBytes += 1

		case Op_i32_eq:
			ansOps = append(ansOps, i32Eq{})
			pointInBytes += 1

		case Op_i32_ne:
			ansOps = append(ansOps, i32Ne{})
			pointInBytes += 1

		case Op_i32_lt_s:
			ansOps = append(ansOps, i32Lts{})
			pointInBytes += 1

		case Op_i32_lt_u:
			ansOps = append(ansOps, i32Ltu{})
			pointInBytes += 1

		case Op_i32_gt_s:
			ansOps = append(ansOps, i32Gts{})
			pointInBytes += 1

		case Op_i32_gt_u:
			ansOps = append(ansOps, i32Gtu{})
			pointInBytes += 1
			
		case Op_i32_le_s:
			ansOps = append(ansOps, i32Les{})
			pointInBytes += 1

		case Op_i32_le_u:
			ansOps = append(ansOps, i32Leu{})
			pointInBytes += 1

		case Op_i32_ge_s:
			ansOps = append(ansOps, i32Ges{})
			pointInBytes += 1

		case Op_i32_ge_u:
			ansOps = append(ansOps, i32Eqz{})
			pointInBytes += 1

		
		case Op_i64_eqz:
			ansOps = append(ansOps, i64Eqz{})
			pointInBytes += 1
		case Op_i64_eq:
			ansOps = append(ansOps, i64Eq{})
			pointInBytes += 1
		case Op_i64_ne:
			ansOps = append(ansOps, i64Ne{})
			pointInBytes += 1
		case Op_i64_le_s:
			ansOps = append(ansOps, i64Les{})
			pointInBytes += 1
		case Op_i64_le_u:
			ansOps = append(ansOps, i64Leu{})
			pointInBytes += 1
		case Op_i64_ge_s:
			ansOps = append(ansOps, i64Ges{})
			pointInBytes += 1
		case Op_i64_ge_u:
			ansOps = append(ansOps, i64Geu{})
			pointInBytes += 1

		case Op_i64_and:
			ansOps = append(ansOps, i64And{})
			pointInBytes += 1
		
		case Op_i64_lt_s:
			ansOps = append(ansOps, i64Lts{})
			pointInBytes += 1

		case Op_i64_lt_u:
			ansOps = append(ansOps, i64Ltu{})
			pointInBytes += 1
		
		case Op_i64_gt_u:
			ansOps = append(ansOps, i64Gtu{})
			pointInBytes += 1
		
		case Op_i64_gt_s:
			ansOps = append(ansOps, i64Gts{})
			pointInBytes += 1

		case Op_i64_clz:
			ansOps = append(ansOps, i64Clz{})
			pointInBytes += 1

		case Op_i64_ctz:
			ansOps = append(ansOps, i64Ctz{})
			pointInBytes += 1

		case Op_i64_popcnt:
			ansOps = append(ansOps, i64PopCnt{})
			pointInBytes += 1

		case Op_i64_or:
			ansOps = append(ansOps, i64Or{})
			pointInBytes += 1
		case Op_i64_xor:
			ansOps = append(ansOps, i64Xor{})
			pointInBytes += 1
		case Op_i64_shl:
			ansOps = append(ansOps, i64Shl{})
			pointInBytes += 1
		case Op_i64_shr_s:
			ansOps = append(ansOps, i64Shrs{})
			pointInBytes += 1
		case Op_i64_shr_u:
			ansOps = append(ansOps, i64Shru{})
			pointInBytes += 1
		case Op_i64_rotl:
			ansOps = append(ansOps, i64Rotl{})
			pointInBytes += 1
		case Op_i64_rotr:
			ansOps = append(ansOps, i64Rotr{})
			pointInBytes += 1

		case Op_i64_const:
			var op = i64Const{int64(bytes[pointInBytes+1])}
			ansOps = append(ansOps, op)
			pointInBytes += 2
		case Op_i64_add:
			ansOps = append(ansOps, i64Add{})
			pointInBytes += 1
		case Op_i64_sub:
			ansOps = append(ansOps, i64Sub{})
			pointInBytes += 1
		case Op_i64_mul:
			ansOps = append(ansOps, i64Mul{})
			pointInBytes += 1
		case Op_i64_div_s:
			ansOps = append(ansOps, i64Divs{})
			pointInBytes += 1
		case Op_i64_div_u:
			ansOps = append(ansOps, i64Divu{})
			pointInBytes += 1
		
		case Op_i64_rem_s:
			ansOps = append(ansOps, i64Rems{})
			pointInBytes += 1

		case Op_i64_rem_u:
			ansOps = append(ansOps, i64Remu{})
			pointInBytes += 1

		case Op_if:
			//has a param block type, im not sure what its used for, so lets ignore that...
			//the rest of the conditional statements code must be filled at the end
			ansOps = append(ansOps, opIf{0, 0})
			pointInBytes += 2
		case Op_else:
			//have to look back for the last if statement that does not yet have an end or an else statement
			foundIf := false
			for i := len(ansOps) - 1; i >= 0 && !foundIf; i-- {
				if reflect.TypeOf(ansOps[i]) == reflect.TypeOf(opIf{}) {
					foo := ansOps[i].(opIf)
					if foo.elsePoint == 0 && foo.endPoint == 0 {
						foo.elsePoint = int64(len(ansOps))
						ansOps[i] = foo
						foundIf = true
					}
				}
			}
			pointInBytes += 1
		case Op_end:
			//look for last condition statement
			//TODO: add check for loop, and block statements
			foundConditional := false
			for i := len(ansOps) - 1; i >= 0 && !foundConditional; i-- {
				if reflect.TypeOf(ansOps[i]) == reflect.TypeOf(opIf{}) {
					foo := ansOps[i].(opIf)
					if foo.endPoint == 0 {
						foo.endPoint = int64(len(ansOps))
						ansOps[i] = foo
						foundConditional = true
					}
				}
			}
			pointInBytes += 1
		case Op_current_memory:
			ansOps = append(ansOps, currentMemory{})
			pointInBytes += 1
		case Op_grow_memory:
			ansOps = append(ansOps, growMemory{})
			pointInBytes += 1

		case Op_call:
			pointInBytes += 1
			ansOps = append(ansOps, call{bytes[pointInBytes : pointInBytes+64]})
			pointInBytes += 64
		default:
			print("skipping over byte at: ")
			println(pointInBytes)
			print("with value: ")
			println(hex.EncodeToString(bytes[pointInBytes : pointInBytes+1]))
			pointInBytes += 1
		}

	}

	return ansOps
}
