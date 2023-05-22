package VM

import (
	"errors"
)

var (
	ErrOutOfGas                 = errors.New("out of gas")
	ErrCodeStoreOutOfGas        = errors.New("contract creation code storage out of gas")
	ErrDepth                    = errors.New("max call depth exceeded")
	ErrStackConsistency         = errors.New("inconsistent stack state after control flow exectution")
	ErrInvalidBr                = errors.New("invalid jump destination")
	ErrIfTopElementOfStack      = errors.New("invalid Operand retrieved from stack - Op_If expected")
	ErrInsufficientBalance      = errors.New("insufficient balance for transfer")
	ErrExecutionReverted        = errors.New("execution reverted")
	ErrNonceUintOverflow        = errors.New("nonce overflow")
	ErrContractAddressCollision = errors.New("contract address collision")
	ErrMaxCodeSizeExceeded      = errors.New("max code size exceeded")

	//offchain DB interaction errors
	ErrConnectionRefused = errors.New("403 Forbidden. Server could not be connected to")
)
