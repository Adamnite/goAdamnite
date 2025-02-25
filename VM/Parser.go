package VM

import (
	"bytes"
	"encoding/hex"
	"math"
)

var reader = bytes.NewReader

func parseBytes(bytes []byte) ([]OperationCommon, []ControlBlock) {
	ansOps := []OperationCommon{}
	pointInBytes := 0

	// The first control here marks the beginning of the function
	controlBlocks := []ControlBlock{{
		startAt: 0,
		op:      0x0,
	}}

	index := 0

	for pointInBytes < len(bytes) {
		switch bytes[pointInBytes] {

		case Op_i32_const:
			num, count, err := DecodeInt32(reader(bytes[pointInBytes+1:]))
			if err != nil {
				panic("Error parsing Op_i32_const value")
			}
			ansOps = append(ansOps, i32Const{int32(num), GasQuickStep})
			pointInBytes += int(count) + 1

		case Op_i32_sub:
			ansOps = append(ansOps, i32Sub{GasQuickStep})
			pointInBytes += 1
		case Op_i32_add:
			ansOps = append(ansOps, i32Add{GasQuickStep})
			pointInBytes += 1
		case Op_i32_div_s:
			ansOps = append(ansOps, i32Divs{GasQuickStep})
			pointInBytes += 1
		case Op_i32_clz:
			ansOps = append(ansOps, i32Clz{GasQuickStep})
			pointInBytes += 1
		case Op_i32_ctz:
			ansOps = append(ansOps, i32Ctz{GasQuickStep})
			pointInBytes += 1

		case Op_i32_popcnt:
			ansOps = append(ansOps, i32PopCnt{GasQuickStep})
			pointInBytes += 1
		case Op_i32_mul:
			ansOps = append(ansOps, i32Mul{GasQuickStep})
			pointInBytes += 1

		case Op_i32_rem_s:
			ansOps = append(ansOps, i32Rems{GasQuickStep})
			pointInBytes += 1

		case Op_i32_rem_u:
			ansOps = append(ansOps, i32Remu{GasQuickStep})
			pointInBytes += 1

		case Op_i32_and:
			ansOps = append(ansOps, i32And{GasQuickStep})
			pointInBytes += 1

		case Op_i32_or:
			ansOps = append(ansOps, i32Or{GasQuickStep})
			pointInBytes += 1
		case Op_i32_xor:
			ansOps = append(ansOps, i32Xor{GasQuickStep})
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
			num, count, err := DecodeInt64(reader(bytes[pointInBytes+1:]))

			if err != nil {
				panic("Error parsing Op_i64_const value")
			}
			ansOps = append(ansOps, i64Const{int64(num), GasQuickStep})
			pointInBytes += int(count) + 1

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

		case Op_block:
			controlBlock := ControlBlock{}
			// The next byte is be the block signature(aka blocktype) when it's 0x40 it means empty signature
			controlBlock.signature = bytes[pointInBytes+1]
			controlBlock.op = Op_block
			pointInBytes++

			controlBlock.startAt = uint64(len(ansOps))
			controlBlocks = append(controlBlocks, controlBlock)
			ansOps = append(ansOps, Block{
				index: uint32(len(controlBlocks)) - 1,
				gas:   GasQuickStep,
			})
			pointInBytes++

		case Op_br:
			label, count, err := DecodeUint32(reader(bytes[pointInBytes+1:]))
			if err != nil {
				panic("Error occured while parsing label Op_br")
			}
			ansOps = append(ansOps, Br{
				index: label,
				gas:   GasQuickStep,
			})
			pointInBytes += int(count) + 1

		case Op_br_if:
			label, count, err := DecodeUint32(reader(bytes[pointInBytes+1:]))

			if err != nil {
				panic("Error occured while parsing label Op_brIf")
			}

			ansOps = append(ansOps, BrIf{
				index: label,
				gas:   GasQuickStep,
			})
			pointInBytes += int(count) + 1

		case Op_if:
			controlBlock := ControlBlock{}
			controlBlock.signature = bytes[pointInBytes+1]
			controlBlock.op = Op_if
			pointInBytes++

			controlBlock.startAt = uint64(len(ansOps))
			controlBlocks = append(controlBlocks, controlBlock)
			ansOps = append(ansOps, If{
				index: uint32(len(controlBlocks)) - 1,
				gas:   GasQuickStep,
			})
			pointInBytes++

		case Op_else:
			ifblock := &controlBlocks[len(controlBlocks)-1]

			if ifblock.op != Op_if {
				panic("Last control block element is not an if")
			}

			ifblock.elseAt = uint64(len(ansOps))
			ansOps = append(ansOps, Else{
				index: uint32(ifblock.startAt),
				gas:   GasQuickStep,
			})
			pointInBytes++

		case Op_loop:
			controlBlock := ControlBlock{}
			controlBlock.signature = bytes[pointInBytes+1]
			pointInBytes++

			controlBlock.op = Op_loop
			controlBlock.startAt = uint64(len(ansOps))

			controlBlocks = append(controlBlocks, controlBlock)

			ansOps = append(ansOps, Loop{uint32(len(controlBlocks)) - 1, GasQuickStep})
			pointInBytes++

		case Op_end:
			// Retrieve the block for which we found the end
			block := &controlBlocks[len(controlBlocks)-index-1]

			ansOps = append(ansOps, End{})
			pointInBytes += 1

			block.endAt = uint64(len(ansOps)) - 1
			index++

		case Op_return:
			ansOps = append(ansOps, Return{})
			pointInBytes++

		case Op_call:
			funcIndex, count, err := DecodeUint32(reader(bytes[pointInBytes+1:]))

			if err != nil {
				panic("Error occurred while parsing label Op_call")
			}

			ansOps = append(ansOps, Call{funcIndex, GasFastStep})
			pointInBytes += int(count) + 1

		case Op_call_indirect:
			ansOps = append(ansOps, CallIndirect{})
			pointInBytes++

		case Op_get_local:
			index, count, err := DecodeUint32(reader(bytes[pointInBytes+1:]))

			if err != nil {
				panic("Error occurred while parsing label Op_get_local")
			}

			ansOps = append(ansOps, localGet{int64(index), GasQuickStep})
			pointInBytes += int(count) + 1
		case Op_set_local:
			index, count, err := DecodeUint32(reader(bytes[pointInBytes+1:]))

			if err != nil {
				panic("Error occurred while parsing label Op_set_local")
			}

			ansOps = append(ansOps, localSet{int64(index), GasQuickStep})
			pointInBytes += int(count) + 1

		case Op_get_global:
			index, count, err := DecodeUint32(reader(bytes[pointInBytes+1:]))

			if err != nil {
				panic("Error occurred while parsing label Op_get_global")
			}

			ansOps = append(ansOps, GlobalGet{int64(index), GasQuickStep})
			pointInBytes += int(count) + 1
		case Op_set_global:
			index, count, err := DecodeUint32(reader(bytes[pointInBytes+1:]))

			if err != nil {
				panic("Error occurred while parsing label Op_set_global")
			}

			ansOps = append(ansOps, GlobalSet{int64(index), GasQuickStep})
			pointInBytes += int(count) + 1
		case Op_drop:
			ansOps = append(ansOps, Drop{})
			pointInBytes++
		case Op_select:
			pointInBytes++
		case Op_current_memory:
			ansOps = append(ansOps, currentMemory{})
			pointInBytes += 1
		case Op_grow_memory:
			ansOps = append(ansOps, growMemory{})
			pointInBytes += 1

		case Op_tee_local:
			ansOps = append(ansOps, TeeLocal{uint64(bytes[pointInBytes+1]), GasQuickStep})
			pointInBytes += 2

		case Op_nop:
			ansOps = append(ansOps, NoOp{})
			pointInBytes += 1
		case Op_unreachable:
			ansOps = append(ansOps, UnReachable{})
			pointInBytes += 1

		case Op_i32_wrap_i64:
			ansOps = append(ansOps, i32Wrapi64{})
			pointInBytes += 1
		case Op_i32_trunc_s_f32, Op_i32_trunc_u_f32:
			ansOps = append(ansOps, i32Truncsf32{})
			pointInBytes += 1
		case Op_i32_trunc_s_f64, Op_i32_trunc_u_f64:
			ansOps = append(ansOps, i32Truncsf64{})
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
			num := LE.Uint32(bytes[pointInBytes+1 : 4])
			ansOps = append(ansOps, f32Const{math.Float32frombits(num), GasQuickStep})
			pointInBytes += 5
		case Op_f32_eq:
			ansOps = append(ansOps, f32Eq{GasQuickStep})
			pointInBytes += 1
		case Op_f32_ne:
			ansOps = append(ansOps, f32Neq{GasQuickStep})
			pointInBytes += 1
		case Op_f32_lt:
			ansOps = append(ansOps, f32Lt{GasQuickStep})
			pointInBytes += 1
		case Op_f32_gt:
			ansOps = append(ansOps, f32Gt{GasQuickStep})
			pointInBytes += 1
		case Op_f32_le:
			ansOps = append(ansOps, f32Le{GasQuickStep})
			pointInBytes += 1
		case Op_f32_ge:
			ansOps = append(ansOps, f32Ge{GasQuickStep})
			pointInBytes += 1
		case Op_f32_abs:
			ansOps = append(ansOps, f32Abs{GasQuickStep})
			pointInBytes += 1
		case Op_f32_neg:
			ansOps = append(ansOps, f32Neg{GasQuickStep})
			pointInBytes += 1
		case Op_f32_ceil:
			ansOps = append(ansOps, f32Ceil{GasQuickStep})
			pointInBytes += 1
		case Op_f32_floor:
			ansOps = append(ansOps, f32Floor{GasQuickStep})
			pointInBytes += 1
		case Op_f32_trunc:
			ansOps = append(ansOps, f32Trunc{GasQuickStep})
			pointInBytes += 1
		case Op_f32_nearest:
			ansOps = append(ansOps, f32Nearest{GasQuickStep})
			pointInBytes += 1
		case Op_f32_sqrt:
			ansOps = append(ansOps, f32Sqrt{GasQuickStep})
			pointInBytes += 1
		case Op_f32_add:
			ansOps = append(ansOps, f32Add{GasQuickStep})
			pointInBytes += 1
		case Op_f32_sub:
			ansOps = append(ansOps, f32Sub{GasQuickStep})
			pointInBytes += 1
		case Op_f32_mul:
			ansOps = append(ansOps, f32Mul{GasQuickStep})
			pointInBytes += 1
		case Op_f32_div:
			ansOps = append(ansOps, f32Div{GasQuickStep})
			pointInBytes += 1
		case Op_f32_min:
			ansOps = append(ansOps, f32Min{GasQuickStep})
			pointInBytes += 1
		case Op_f32_max:
			ansOps = append(ansOps, f32Max{GasQuickStep})
			pointInBytes += 1
		case Op_f32_copysign:
			ansOps = append(ansOps, f32CopySign{GasQuickStep})
			pointInBytes += 1

		case Op_f64_const:
			num := LE.Uint64(bytes[pointInBytes+1:])
			ansOps = append(ansOps, f64Const{math.Float64frombits(num), GasQuickStep})
			pointInBytes += 9

		case Op_f64_eq:
			ansOps = append(ansOps, f64Eq{GasQuickStep})
			pointInBytes += 1
		case Op_f64_ne:
			ansOps = append(ansOps, f64Ne{GasQuickStep})
			pointInBytes += 1
		case Op_f64_lt:
			ansOps = append(ansOps, f64Lt{GasQuickStep})
			pointInBytes += 1
		case Op_f64_gt:
			ansOps = append(ansOps, f64Gt{GasQuickStep})
			pointInBytes += 1
		case Op_f64_le:
			ansOps = append(ansOps, f64Le{GasQuickStep})
			pointInBytes += 1
		case Op_f64_ge:
			ansOps = append(ansOps, f64Ge{GasQuickStep})
			pointInBytes += 1
		case Op_f64_abs:
			ansOps = append(ansOps, f64Abs{GasQuickStep})
			pointInBytes += 1
		case Op_f64_neg:
			ansOps = append(ansOps, f64Neg{GasQuickStep})
			pointInBytes += 1
		case Op_f64_ceil:
			ansOps = append(ansOps, f64Ceil{GasQuickStep})
			pointInBytes += 1
		case Op_f64_floor:
			ansOps = append(ansOps, f64Floor{GasQuickStep})
			pointInBytes += 1
		case Op_f64_trunc:
			ansOps = append(ansOps, f64Trunc{GasQuickStep})
			pointInBytes += 1
		case Op_f64_nearest:
			ansOps = append(ansOps, f64Nearest{GasQuickStep})
			pointInBytes += 1
		case Op_f64_sqrt:
			ansOps = append(ansOps, f64Sqrt{GasQuickStep})
			pointInBytes += 1
		case Op_f64_add:
			ansOps = append(ansOps, f64Add{GasQuickStep})
			pointInBytes += 1
		case Op_f64_sub:
			ansOps = append(ansOps, f64Sub{GasQuickStep})
			pointInBytes += 1
		case Op_f64_mul:
			ansOps = append(ansOps, f64Mul{GasQuickStep})
			pointInBytes += 1
		case Op_f64_div:
			ansOps = append(ansOps, f64Div{GasQuickStep})
			pointInBytes += 1
		case Op_f64_min:
			ansOps = append(ansOps, f64Min{GasQuickStep})
			pointInBytes += 1
		case Op_f64_max:
			ansOps = append(ansOps, f64Max{GasQuickStep})
			pointInBytes += 1
		case Op_f64_copysign:
			ansOps = append(ansOps, f64CopySign{GasQuickStep})
			pointInBytes += 1

		case Op_i32_load, Op_i64_load32_u:
			ansOps = append(ansOps, i32Load{
				align:  uint32(bytes[pointInBytes+1]),
				offset: uint32(bytes[pointInBytes+2]),
				gas:    GasQuickStep,
			})
			pointInBytes += 3

		case Op_i32_store, Op_i64_store32:
			ansOps = append(ansOps, i32Store{
				align:  uint32(bytes[pointInBytes+1]),
				offset: uint32(bytes[pointInBytes+2]),
				gas:    GasQuickStep,
			})
			pointInBytes += 3

		case Op_i64_load:
			ansOps = append(ansOps, i64Load{
				align:  uint32(bytes[pointInBytes+1]),
				offset: uint32(bytes[pointInBytes+2]),
				gas:    GasQuickStep,
			})
			pointInBytes += 3
		case Op_i64_store:
			ansOps = append(ansOps, i64Store{
				align:  uint32(bytes[pointInBytes+1]),
				offset: uint32(bytes[pointInBytes+2]),
				gas:    GasQuickStep,
			})
			pointInBytes += 3
		case Op_i32_load8_s, Op_i64_load8_s:
			ansOps = append(ansOps, i32Load8s{
				align:  uint32(bytes[pointInBytes+1]),
				offset: uint32(bytes[pointInBytes+2]),
				gas:    GasQuickStep,
			})
			pointInBytes += 3
		case Op_i32_store8, Op_i64_store8:
			ansOps = append(ansOps, i32Store8{
				align:  uint32(bytes[pointInBytes+1]),
				offset: uint32(bytes[pointInBytes+2]),
				gas:    GasQuickStep,
			})
			pointInBytes += 3
		case Op_i32_load8_u, Op_i64_load8_u:
			ansOps = append(ansOps, i32Load8u{
				align:  uint32(bytes[pointInBytes+1]),
				offset: uint32(bytes[pointInBytes+2]),
				gas:    GasQuickStep,
			})

			pointInBytes += 3

		case Op_i64_load32_s:
			ansOps = append(ansOps, i64Load32s{
				align:  uint32(bytes[pointInBytes+1]),
				offset: uint32(bytes[pointInBytes+2]),
				gas:    GasQuickStep,
			})

			pointInBytes += 3

		case Op_i32_load16_u, Op_i64_load16_u:
			ansOps = append(ansOps, i32Load16u{
				align:  uint32(bytes[pointInBytes+1]),
				offset: uint32(bytes[pointInBytes+2]),
				gas:    GasQuickStep,
			})

			pointInBytes += 3

		case Op_i64_load16_s, Op_i32_load16_s:
			ansOps = append(ansOps, i64Load16s{
				align:  uint32(bytes[pointInBytes+1]),
				offset: uint32(bytes[pointInBytes+2]),
				gas:    GasQuickStep,
			})

			pointInBytes += 3

		case Op_i32_store16, Op_i64_store16:
			ansOps = append(ansOps, i32Store16{
				align:  uint32(bytes[pointInBytes+1]),
				offset: uint32(bytes[pointInBytes+2]),
				gas:    GasQuickStep,
			})

			pointInBytes += 3

		case Op_address:
			ansOps = append(ansOps, opAddress{GasQuickStep})
			pointInBytes++
		case Op_balance:
			ansOps = append(ansOps, balance{GasQuickStep})
			pointInBytes++

		case Op_timestamp:
			ansOps = append(ansOps, blocktimestamp{GasQuickStep})
			pointInBytes++
		case Op_value:
			ansOps = append(ansOps, valueOp{GasQuickStep})
			pointInBytes++
		case Op_data_size:
			ansOps = append(ansOps, dataSize{GasQuickStep})
			pointInBytes++
		case Op_caller:
			ansOps = append(ansOps, callerAddr{GasQuickStep})
			pointInBytes++
		case Op_get_data:
			ansOps = append(ansOps, getData{GasQuickStep})
			pointInBytes++
		case Op_get_code:
			ansOps = append(ansOps, getCode{GasQuickStep})
			pointInBytes++

		case Op_copy_code:
			ansOps = append(ansOps, copyCode{GasQuickStep})
			pointInBytes++

		default:
			print("skipping over byte at: ")
			println(pointInBytes)
			print("with value: ")
			println(hex.EncodeToString(bytes[pointInBytes : pointInBytes+1]))
			pointInBytes += 1
		}

	}

	return ansOps, controlBlocks
}
