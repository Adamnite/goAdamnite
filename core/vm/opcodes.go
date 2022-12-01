package vm

type Opcode byte

// type OpSize uint8

// Directly from WASM, and from the Adamnite Technical Paper and Document

// Language types opcodes as defined by:
// http://webassembly_org/docs/binary-encoding/#language-types
const (
	Op_i32     = 0x7f
	Op_i64     = 0x7e
	Op_f32     = 0x7d
	Op_f64     = 0x7c
	Op_anyfunc = 0x70
	Op_func    = 0x60
	Op_empty   = 0x40
)

// Control flow operators
const (
	Op_unreachable = 0x00 // done
	Op_nop         = 0x01 // done
	Op_block       = 0x02
	Op_loop        = 0x03
	Op_if          = 0x04 //
	Op_else        = 0x05 //
	Op_end         = 0x0b
	Op_br          = 0x0c
	Op_br_if       = 0x0d
	Op_br_table    = 0x0e
	Op_return      = 0x0f
)

// Call operators
const (
	Op_call          = 0x10
	Op_call_indirect = 0x11
	Op_delegate_call = 0x18
)

// Parametric operators
const (
	Op_drop   = 0x1a // done
	Op_select = 0x1b // done
)

// Variable access
const (
	Op_get_local  = 0x20 //
	Op_set_local  = 0x21 //
	Op_tee_local  = 0x22
	Op_get_global = 0x23
	Op_set_global = 0x24
)

// Memory-related operators
const (
	Op_i32_load    = 0x28 // done
	Op_i64_load    = 0x29 // done
	Op_f32_load    = 0x2a
	Op_f64_load    = 0x2b
	Op_i32_load8_s = 0x2c // done
	Op_i32_load8_u = 0x2d // done

	Op_i32_load16_s = 0x2e // done
	Op_i32_load16_u = 0x2f // done

	Op_i64_load8_s  = 0x30 // done
	Op_i64_load8_u  = 0x31 // done
	Op_i64_load16_s = 0x32 // done
	Op_i64_load16_u = 0x33 // done
	Op_i64_load32_s = 0x34 // done
	Op_i64_load32_u = 0x35 // done
	Op_i32_store    = 0x36 // done
	Op_i64_store    = 0x37 // done
	Op_f32_store    = 0x38
	Op_f64_store    = 0x39
	Op_i32_store8   = 0x3a // done

	Op_i32_store16 = 0x3b
	Op_i64_store8  = 0x3c
	Op_i64_store16 = 0x3d

	Op_i64_store32    = 0x3e // done
	Op_current_memory = 0x3f //
	Op_grow_memory    = 0x40 //
)

// Constants opcodes
const (
	Op_i32_const = 0x41 // Done
	Op_i64_const = 0x42 //
	Op_f32_const = 0x43 // done
	Op_f64_const = 0x44 // done
)

// Comparison operators
const (
	Op_i32_eqz  = 0x45 // done
	Op_i32_eq   = 0x46 // done
	Op_i32_ne   = 0x47 // done
	Op_i32_lt_s = 0x48 // done
	Op_i32_lt_u = 0x49 // done
	Op_i32_gt_s = 0x4a // done
	Op_i32_gt_u = 0x4b // done
	Op_i32_le_s = 0x4c // done
	Op_i32_le_u = 0x4d // done
	Op_i32_ge_s = 0x4e // done
	Op_i32_ge_u = 0x4f // done

	Op_i64_eqz  = 0x50 //
	Op_i64_eq   = 0x51 //
	Op_i64_ne   = 0x52 //
	Op_i64_lt_s = 0x53 // done
	Op_i64_lt_u = 0x54 // done
	Op_i64_gt_s = 0x55 // done
	Op_i64_gt_u = 0x56 // done
	Op_i64_le_s = 0x57 //
	Op_i64_le_u = 0x58 //
	Op_i64_ge_s = 0x59 //
	Op_i64_ge_u = 0x5a //

	Op_f32_eq = 0x5b // done
	Op_f32_ne = 0x5c // done
	Op_f32_lt = 0x5d // done
	Op_f32_gt = 0x5e // done
	Op_f32_le = 0x5f // done
	Op_f32_ge = 0x60 // done

	Op_f64_eq = 0x61 // done
	Op_f64_ne = 0x62 // done
	Op_f64_lt = 0x63 // done
	Op_f64_gt = 0x64 // done
	Op_f64_le = 0x65 // done
	Op_f64_ge = 0x66 // done
)

// Numeric operators
const (
	Op_i32_clz    = 0x67 // Done
	Op_i32_ctz    = 0x68 // Done
	Op_i32_popcnt = 0x69 // Done
	Op_i32_add    = 0x6a // done
	Op_i32_sub    = 0x6b // done
	Op_i32_mul    = 0x6c // done
	Op_i32_div_s  = 0x6d // Done
	Op_i32_div_u  = 0x6e // done
	Op_i32_rem_s  = 0x6f // Done
	Op_i32_rem_u  = 0x70 // done
	Op_i32_and    = 0x71 // done
	Op_i32_or     = 0x72 // done
	Op_i32_xor    = 0x73 // done
	Op_i32_shl    = 0x74 // done
	Op_i32_shr_s  = 0x75 // done
	Op_i32_shr_u  = 0x76 // done
	Op_i32_rotl   = 0x77 // done
	Op_i32_rotr   = 0x78 // done

	Op_i64_clz      = 0x79 // done
	Op_i64_ctz      = 0x7a // done
	Op_i64_popcnt   = 0x7b // done
	Op_i64_add      = 0x7c //
	Op_i64_sub      = 0x7d //
	Op_i64_mul      = 0x7e //
	Op_i64_div_s    = 0x7f // done
	Op_i64_div_u    = 0x80 // done
	Op_i64_rem_s    = 0x81 // done
	Op_i64_rem_u    = 0x82 // done
	Op_i64_and      = 0x83 //
	Op_i64_or       = 0x84 //
	Op_i64_xor      = 0x85 //
	Op_i64_shl      = 0x86 // done
	Op_i64_shr_s    = 0x87 // done
	Op_i64_shr_u    = 0x88 // done
	Op_i64_rotl     = 0x89 // done
	Op_i64_rotr     = 0x8a // done
	Op_f32_abs      = 0x8b // done
	Op_f32_neg      = 0x8c // done
	Op_f32_ceil     = 0x8d // done
	Op_f32_floor    = 0x8e // done
	Op_f32_trunc    = 0x8f // done
	Op_f32_nearest  = 0x90 // done
	Op_f32_sqrt     = 0x91 // done
	Op_f32_add      = 0x92 // done
	Op_f32_sub      = 0x93 // done
	Op_f32_mul      = 0x94 // done
	Op_f32_div      = 0x95 // done
	Op_f32_min      = 0x96 // done
	Op_f32_max      = 0x97 // done
	Op_f32_copysign = 0x98 // done
	Op_f64_abs      = 0x99 // done
	Op_f64_neg      = 0x9a // done
	Op_f64_ceil     = 0x9b // done
	Op_f64_floor    = 0x9c // done
	Op_f64_trunc    = 0x9d // done
	Op_f64_nearest  = 0x9e // done
	Op_f64_sqrt     = 0x9f // done
	Op_f64_add      = 0xa0 // done
	Op_f64_sub      = 0xa1 // done
	Op_f64_mul      = 0xa2 // done
	Op_f64_div      = 0xa3 // done
	Op_f64_min      = 0xa4 // done
	Op_f64_max      = 0xa5 // done
	Op_f64_copysign = 0xa6 // done
)

// Conversions
const (
	Op_i32_wrap_i64      = 0xa7 // done
	Op_i32_trunc_s_f32   = 0xa8 // done
	Op_i32_trunc_u_f32   = 0xa9 // done
	Op_i32_trunc_s_f64   = 0xaa // done
	Op_i32_trunc_u_f64   = 0xab // done
	Op_i64_extend_s_i32  = 0xac // done
	Op_i64_extend_u_i32  = 0xad // done
	Op_i64_trunc_s_f32   = 0xae // done
	Op_i64_trunc_u_f32   = 0xaf // done
	Op_i64_trunc_s_f64   = 0xb0 // done
	Op_i64_trunc_u_f64   = 0xb1 // done
	Op_f32_convert_s_i32 = 0xb2 // done
	Op_f32_convert_u_i32 = 0xb3 // done
	Op_f32_convert_s_i64 = 0xb4 // done
	Op_f32_convert_u_i64 = 0xb5 // done

	Op_f32_demote_f64    = 0xb6 // done
	Op_f64_convert_s_i32 = 0xb7 // done
	Op_f64_convert_u_i32 = 0xb8 // done
	Op_f64_convert_s_i64 = 0xb9 // done
	Op_f64_convert_u_i64 = 0xba // done
	Op_f64_promote_f32   = 0xbb // done
)

// Environment Related Operations
const (
	Op_address        = 0xc1 //address of the contract
	Op_balance        = 0xc2 //balance of the contract
	Op_caller         = 0xc3 //address of the caller
	Op_datasize       = 0xc4
	Op_caller_balance = 0xc5 //balance of the caller
	Op_timestamp      = 0xc6 //blocks timestamp
)

// Fee and storage level operations
const (
	Op_value     = 0xd1
	Op_gas_price = 0xd2
	Op_code_size = 0xd3
	Op_data_size = 0xd4
	Op_get_code  = 0xd5
	Op_copy_code = 0xd6
	Op_get_data  = 0xd7
)
