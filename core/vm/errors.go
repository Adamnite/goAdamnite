package vm

import (
	"errors")

var (
	ErrOutOfGas                 = errors.New("out of gas")
	ErrCodeStoreOutOfGas        = errors.New("contract creation code storage out of gas")
	ErrDepth                    = errors.New("max call depth exceeded")
	ErrStackConsistency			= errors.New("inconsistent stack state after control flow exectution")
	ErrInvalidBr	            = errors.New("invalid jump destination")
	ErrIfTopElementOfStack		= errors.New("Invalid Operand retrieved from stack - Op_If expected")

)