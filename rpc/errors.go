package rpc

import "errors"

var (
	ErrStateNotSet = errors.New("StateDB was not established")
	ErrChainNotSet = errors.New("chain reference not filled")
)
