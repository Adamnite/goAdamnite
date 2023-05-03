package rpc

import (
	"errors"
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/ugorji/go/codec"
	"github.com/vmihailenco/msgpack/v5"
)

type PassedContacts struct {
	NodeIDs                    []common.Address
	ConnectionStrings          []string
	BlacklistIDs               []common.Address
	BlacklistConnectionStrings []string
}

func Encode(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func Decode(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

type AdmVersionReply struct {
	Client_version string
	Timestamp      time.Time
	Addr_received  common.Address //address is passed as a string
	Addr_from      common.Address
	Last_round     *big.Int
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
	ErrStateNotSet        = errors.New("StateDB was not established")
	ErrChainNotSet        = errors.New("chain reference not filled")
	ErrPreExistingAccount = errors.New("specified account already exists on chain")
	ErrNoAccountSet       = errors.New("the account address has not been set")
)
