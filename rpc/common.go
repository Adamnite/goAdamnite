package rpc

import (
	"math/big"

	"github.com/ugorji/go/codec"
)

type BigIntReply struct {
	Value []byte
}

func (b *BigIntReply) toBigInt() *big.Int {
	return big.NewInt(0).SetBytes(b.Value)
}
func BigIntReplyFromBytes(val []byte) BigIntReply {
	return BigIntReply{Value: val}
}
func BigIntReplyFromBigInt(val big.Int) BigIntReply {
	return BigIntReply{Value: val.Bytes()}
}

var (
	mh codec.MsgpackHandle
	// msgpackHandler = codec.MsgpackHandle{
	// 	NoFixedNum:          true,
	// 	WriteExt:            true,
	// 	PositiveIntUnsigned: false,
	// }
)
