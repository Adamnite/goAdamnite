package rpc

import (
	"errors"
	"math/big"
	"time"

	"github.com/ugorji/go/codec"
)

type PassedContacts struct {
	NodeIDs                    []int
	ConnectionStrings          []string
	BlacklistIDs               []int
	BlacklistConnectionStrings []string
}
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

type AdmVersionReply struct {
	Client_version string
	Timestamp      time.Time
	Addr_received  string //address is passed as a string
	Addr_from      string
	Last_round     BigIntRPC
	Nonce          int //TODO: check what the nonce should be
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
