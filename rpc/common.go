package rpc

import (
	"encoding/base64"
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

// for internal use of converting http request data towards rpc supported structuring.
func decodeBase64(value *string) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(*value)))
	n, err := base64.StdEncoding.Decode(decoded, []byte(*value))
	if err != nil {
		return nil, err
	}
	return decoded[:n], nil
}

// for converting HTTP request data into rpc conforming data
type RPCRequest struct {
	Method string
	Params string
	Id     int
}
