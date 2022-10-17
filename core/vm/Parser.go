package vm

import (
	"encoding/hex"
	"reflect"
)

func parseBytes(bytes []byte) []OperationCommon {
	ansOps := []OperationCommon{}
	pointInBytes := 0

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
			// Store the current position as the startpoint of the if block
			ifBlock := opIf{uint64(pointInBytes), 0, 0, []OperationCommon{}, Op_if}
			// Parse untill we find 0x0b (Op_end) or 0x05 (Op_else)
			foundTerminator := false
			for i := pointInBytes; i < len(ansOps); i++ {
				if reflect.TypeOf(ansOps[i]) == reflect.TypeOf(opEnd{}) {
					// If without else block
					ifBlock.endPoint = uint64(i)
					ifBlock.code = ansOps[pointInBytes:uint64(i)]
					foundTerminator = true
					break
				}

				if reflect.TypeOf(ansOps[i]) == reflect.TypeOf(opElse{}) {
					// If with else block
					ifBlock.elsePoint = uint64(i)
					ifBlock.code = ansOps[pointInBytes:uint64(i)]
					foundTerminator = true
					break
				}
			}

			if !foundTerminator {
				panic("Block type without an end statement Op_if")
			}
			ansOps = append(ansOps, ifBlock)
		case Op_else:
	
			elseBlock := opElse{uint64(pointInBytes), 0, []OperationCommon{}}
			// Parse untill we find 0x0b (Op_end)
			foundTerminator := false
			for i := pointInBytes; i < len(ansOps); i++ {
				if reflect.TypeOf(ansOps[i]) == reflect.TypeOf(opEnd{}) {
					elseBlock.endPoint = uint64(i)
					elseBlock.code = ansOps[pointInBytes:uint64(i)]
					foundTerminator = true
					break
				}
			}
			if !foundTerminator {
				panic("Block type without an end statement Op_else")
			}

			ansOps = append(ansOps, elseBlock)
		case Op_br:
			ansOps = append(ansOps, opBr{})
			pointInBytes += 1
		case Op_br_if:
			ansOps = append(ansOps, opBrIf{})
		case Op_return:
			// pops return value off the stack and returns from the current function.
			ansOps = append(ansOps, opReturn{})
		
		case Op_block:
			// Store the current position as the startpoint of the block
			block := opBlock{uint64(pointInBytes), 0, []OperationCommon{}}
			// Parse untill we find 0x0b (Op_end)
			foundTerminator := false
			for i := pointInBytes; i < len(ansOps); i++ {
				if reflect.TypeOf(ansOps[i]) == reflect.TypeOf(opEnd{}) {
					block.endPoint = uint64(i)
					foundTerminator = true
					break
				}
			}
			if !foundTerminator {
				panic("Block type without an end statement Op_block")
			}
			block.code = ansOps[pointInBytes:block.endPoint]
			ansOps = append(ansOps, block)
			pointInBytes += int(block.endPoint)
		case Op_end:
			ansOps = append(ansOps, opEnd{})
			pointInBytes += 1
		case Op_loop:
			// Store the current position as the startpoint of the block
			loopBlock := opLoop{uint64(pointInBytes), 0, []OperationCommon{}}
			// Parse untill we find 0x0b (Op_end)
			foundTerminator := false
			for i := pointInBytes; i < len(ansOps); i++ {
				if reflect.TypeOf(ansOps[i]) == reflect.TypeOf(opEnd{}) {
					loopBlock.endPoint = uint64(i)
					foundTerminator = true
					break
				}
			}
			if !foundTerminator {
				panic("Block type without an end statement Op_loop")
			}
			loopBlock.code = ansOps[pointInBytes:loopBlock.endPoint]
			ansOps = append(ansOps, loopBlock)
			pointInBytes += int(loopBlock.endPoint)
		
		case Op_get_local:
			ansOps = append(ansOps, localGet{int64(bytes[pointInBytes+1])})
			pointInBytes += 2
		case Op_drop:
			ansOps = append(ansOps, opDrop{})
			pointInBytes++
		case Op_select:
			ansOps = append(ansOps, opSelect{})
			pointInBytes++
		case Op_current_memory:
			ansOps = append(ansOps, currentMemory{})
			pointInBytes += 1
		case Op_grow_memory:
			ansOps = append(ansOps, growMemory{})
			pointInBytes += 1

		// case Op_call:
		// 	pointInBytes += 1
		// 	ansOps = append(ansOps, call{bytes[pointInBytes : pointInBytes+64]})
		// 	pointInBytes += 64

		case Op_tee_local:
			ansOps = append(ansOps, TeeLocal{uint64(bytes[pointInBytes+1])})
			pointInBytes += 2
		
		case Op_nop:
			ansOps = append(ansOps, noOp{})
			pointInBytes += 1
		case Op_unreachable:
			ansOps = append(ansOps, unReachable{})
			pointInBytes += 1

		case Op_i32_wrap_i64:
			ansOps = append(ansOps, i32Wrapi64{})
			pointInBytes += 1	
		case Op_i32_trunc_s_f32, Op_i32_trunc_u_f32:
			ansOps = append(ansOps, i32Truncsf32{})
			pointInBytes += 1
		case Op_i32_trunc_s_f64, Op_i32_trunc_u_f64:
			ansOps = append(ansOps, i32Truncsf64{});
			pointInBytes += 1
		
		case Op_i64_extend_s_i32:
			ansOps = append(ansOps, i64Extendsi32{})
			pointInBytes += 1
		
		case Op_i64_trunc_s_f32, Op_i64_trunc_u_f32:
			ansOps = append(ansOps, i64Truncsf32{})
			pointInBytes += 1
		
		case Op_i64_trunc_s_f64, Op_i64_trunc_u_f64:
			ansOps = append(ansOps, i64Truncsf64{})
			pointInBytes += 1
		
		case Op_f32_convert_s_i32:
			ansOps = append(ansOps, f32Convertsi32{})
			pointInBytes += 1
		
		case Op_f32_convert_u_i32:
			ansOps = append(ansOps, f32Convertui32{})
			pointInBytes += 1
		
		case Op_i64_extend_u_i32:
			ansOps = append(ansOps, i64Extendui32{})
			pointInBytes += 1
		
		case Op_f32_convert_s_i64:
			ansOps = append(ansOps, f32Convertsi64{})
			pointInBytes += 1

		case Op_f32_convert_u_i64:
			ansOps = append(ansOps, f32Convertui64{})
			pointInBytes += 1
		
		case Op_f32_demote_f64:
			ansOps = append(ansOps, f32Demotef64{})
			pointInBytes += 1
		case Op_f64_convert_s_i32:
			ansOps = append(ansOps, f64convertsi32{})
			pointInBytes += 1
		case Op_f64_convert_u_i32:
			ansOps = append(ansOps, f64convertui32{})
			pointInBytes += 1
		case Op_f64_convert_s_i64:
			ansOps = append(ansOps, f64Convertsi64{})
			pointInBytes += 1
		case Op_f64_convert_u_i64:
			ansOps = append(ansOps, f64Convertui64{})
			pointInBytes += 1
		case Op_f64_promote_f32:
			ansOps = append(ansOps, f64Promotef32{})
			pointInBytes += 1
		
		case Op_f32_const:
			ansOps = append(ansOps, f32Const{float32(bytes[pointInBytes+1])})
			pointInBytes += 2

		case Op_f32_eq:
			ansOps = append(ansOps, f32Eq{})
			pointInBytes +=1
		case Op_f32_ne:
			ansOps = append(ansOps, f32Neq{})
			pointInBytes +=1
		case Op_f32_lt:
			ansOps = append(ansOps, f32Lt{})
			pointInBytes +=1
		case Op_f32_gt:
			ansOps = append(ansOps, f32Gt{})
			pointInBytes +=1
		case Op_f32_le:
			ansOps = append(ansOps, f32Le{})
			pointInBytes +=1
		case Op_f32_ge:
			ansOps = append(ansOps, f32Ge{})
			pointInBytes +=1
		case Op_f32_abs:
			ansOps = append(ansOps, f32Abs{})
			pointInBytes += 1
		case Op_f32_neg:
			ansOps = append(ansOps, f32Neg{})
			pointInBytes += 1
		case Op_f32_ceil:
			ansOps = append(ansOps, f32Ceil{})
			pointInBytes += 1
		case Op_f32_floor:
			ansOps = append(ansOps, f32Floor{})
			pointInBytes += 1
		case Op_f32_trunc:
			ansOps = append(ansOps, f32Trunc{})
			pointInBytes += 1
		case Op_f32_nearest:
			ansOps = append(ansOps, f32Nearest{})
			pointInBytes += 1
		case Op_f32_sqrt:
			ansOps = append(ansOps, f32Sqrt{})
			pointInBytes += 1
		case Op_f32_add:
			ansOps = append(ansOps, f32Add{})
			pointInBytes += 1
		case Op_f32_sub:
			ansOps = append(ansOps, f32Sub{})
			pointInBytes += 1
		case Op_f32_mul:
			ansOps = append(ansOps, f32Mul{})
			pointInBytes += 1
		case Op_f32_div:
			ansOps = append(ansOps, f32Div{})
			pointInBytes += 1
		case Op_f32_min:
			ansOps = append(ansOps, f32Min{})
			pointInBytes += 1
		case Op_f32_max:
			ansOps = append(ansOps, f32Max{})
			pointInBytes += 1
		case Op_f32_copysign:
			ansOps = append(ansOps, f32CopySign{})
			pointInBytes += 1
		
		case Op_f64_const:
			ansOps = append(ansOps, f64Const{float64(bytes[pointInBytes+1])})
			pointInBytes += 2

		case Op_f64_eq:
			ansOps = append(ansOps, f64Eq{})
			pointInBytes += 1
		case Op_f64_ne:
			ansOps = append(ansOps, f64Ne{})
			pointInBytes += 1
		case Op_f64_lt:
			ansOps = append(ansOps, f64Lt{})
			pointInBytes += 1
		case Op_f64_gt:
			ansOps = append(ansOps, f64Gt{})
			pointInBytes += 1
		case Op_f64_le:
			ansOps = append(ansOps, f64Le{})
			pointInBytes += 1
		case Op_f64_ge:
			ansOps = append(ansOps, f64Ge{})
			pointInBytes += 1
		case Op_f64_abs:
			ansOps = append(ansOps, f64Abs{})
			pointInBytes += 1 
		case Op_f64_neg:
			ansOps = append(ansOps, f64Neg{})
			pointInBytes += 1 
		case Op_f64_ceil:
			ansOps = append(ansOps, f64Ceil{})
			pointInBytes += 1 
		case Op_f64_floor:
			ansOps = append(ansOps, f64Floor{})
			pointInBytes += 1 
		case Op_f64_trunc:
			ansOps = append(ansOps, f64Trunc{})
			pointInBytes += 1 
		case Op_f64_nearest:
			ansOps = append(ansOps, f64Nearest{})
			pointInBytes += 1 
		case Op_f64_sqrt:
			ansOps = append(ansOps, f64Sqrt{})
			pointInBytes += 1 
		case Op_f64_add:
			ansOps = append(ansOps, f64Add{})
			pointInBytes += 1 
		case Op_f64_sub:
			ansOps = append(ansOps, f64Sub{})
			pointInBytes += 1 
		case Op_f64_mul:
			ansOps = append(ansOps, f64Mul{})
			pointInBytes += 1 
		case Op_f64_div:
			ansOps = append(ansOps, f64Div{})
			pointInBytes += 1 
		case Op_f64_min:
			ansOps = append(ansOps, f64Min{})
			pointInBytes += 1 
		case Op_f64_max     :
			ansOps = append(ansOps, f64Max{})
			pointInBytes += 1 
		case Op_f64_copysign:
			ansOps = append(ansOps, f64CopySign{})
			pointInBytes += 1
		
		case Op_i32_load, Op_i64_load32_u:
			ansOps = append(ansOps, i32Load{uint32(bytes[pointInBytes+1]), uint32(bytes[pointInBytes+2])})
			pointInBytes += 3
		
		case Op_i32_store, Op_i64_store32:
			ansOps = append(ansOps, i32Store{uint32(bytes[pointInBytes+1]), uint32(bytes[pointInBytes+2])})
			pointInBytes += 3
		
		case Op_i64_load:
			ansOps = append(ansOps, i64Load{uint32(bytes[pointInBytes+1]), uint32(bytes[pointInBytes+2])})
			pointInBytes += 3
		case Op_i64_store:
			ansOps = append(ansOps, i64Store{uint32(bytes[pointInBytes+1]), uint32(bytes[pointInBytes+2])})
			pointInBytes += 3
		case Op_i32_load8_s, Op_i64_load8_s:
			ansOps = append(ansOps, i32Load8s{uint32(bytes[pointInBytes+1]), uint32(bytes[pointInBytes+2])})
			pointInBytes += 3
		case Op_i32_store8, Op_i64_store8:
			ansOps = append(ansOps, i32Store8{uint32(bytes[pointInBytes+1]), uint32(bytes[pointInBytes+2])})
			pointInBytes += 3
		case Op_i32_load8_u, Op_i64_load8_u:
			ansOps = append(ansOps, i32Load8u{uint32(bytes[pointInBytes+1]), uint32(bytes[pointInBytes+2])})
			pointInBytes += 3
		
		case Op_i64_load32_s:
			ansOps = append(ansOps, i64Load32s{uint32(bytes[pointInBytes+1]), uint32(bytes[pointInBytes+2])})
			pointInBytes += 3

		case Op_i32_load16_u, Op_i64_load16_u:
			ansOps = append(ansOps, i32Load16u{uint32(bytes[pointInBytes+1]), uint32(bytes[pointInBytes+2])})
			pointInBytes += 3

		case Op_i64_load16_s, Op_i32_load16_s:
			ansOps = append(ansOps, i64Load16s{uint32(bytes[pointInBytes+1]), uint32(bytes[pointInBytes+2])})
			pointInBytes += 3

		case Op_i32_store16, Op_i64_store16:
			ansOps = append(ansOps, i32Store16{uint32(bytes[pointInBytes+1]), uint32(bytes[pointInBytes+2])})
			pointInBytes += 3
	
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
