//Copyright 2022 The goAdamnite authors

package params

//Protocol Parameters for general transaction, consensus, and execution related parameters
// For VM operation fees, all operations not mentioned here should be 1 dubnite . All operations are envisioned as i64 operations, but can be implemented on i32 with the same fees.

const (
	Ate_Division_Limit      uint64 = 512
	MinimumFee              uint64 = 5
	MaxmiumFeeLimit         uint64 = 0x7fffffffffffffff
	Extra_Data              uint64 = 32
	Application_Call        uint64 = 500  //Paid when one contract calls another contract
	Application_Create_call uint64 = 1200 //Paid when one contract's execution creates a new contract, and calls it
	Transaction_Fee         uint64 = 300  //Base transaction fee
	Contract_Creation_Fee   uint64 = 2000 //Contract Creation Fee
	Mul_Fee                 uint64 = 2
	Div_Fee                 uint64 = 2
	i32load_fee             uint64 = 4
	i64load_fee             uint64 = 8
	i32store_fee            uint64 = 6
	i64store_fee            uint64 = 12
	Sha512_fee              uint64 = 15
	Sha512_fee_per_word     uint64 = 10
	code_size_fee           uint64 = 5
	data_size_fee           uint64 = 5
	code_copy_fee           uint64 = 2
	data_copy_fee           uint64 = 2
	module_fee              uint64 = 10 //Fees for using predefined WASM Modules and enivironments.
)
