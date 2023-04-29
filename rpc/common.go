package rpc

import (
	"math/big"

	"github.com/vmihailenco/msgpack/v5"
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

func Encode(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func Decode(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}
