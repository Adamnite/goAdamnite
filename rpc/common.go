package rpc

import (
	"errors"
	"math/big"

	"github.com/ugorji/go/codec"
)

type BigIntRPC struct {
	Value []byte
}

func (b *BigIntRPC) toBigInt() *big.Int {
	return big.NewInt(0).SetBytes(b.Value)
}
func BigIntReplyFromBytes(val []byte) BigIntRPC {
	return BigIntRPC{Value: val}
}
func BigIntReplyFromBigInt(val big.Int) BigIntRPC {
	return BigIntRPC{Value: val.Bytes()}
}

var (
	mh codec.MsgpackHandle
	// msgpackHandler = codec.MsgpackHandle{
	// 	NoFixedNum:          true,
	// 	WriteExt:            true,
	// 	PositiveIntUnsigned: false,
	// }
)

var (
	ErrStateNotSet = errors.New("StateDB was not established")
	ErrChainNotSet = errors.New("chain reference not filled")
)
