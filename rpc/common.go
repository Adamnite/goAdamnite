package rpc

import (
	"errors"
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/ugorji/go/codec"
	"github.com/vmihailenco/msgpack/v5"
)

type PassedContacts struct {
	NodeIDs                    []common.Address
	ConnectionStrings          []string
	BlacklistIDs               []common.Address
	BlacklistConnectionStrings []string
}
type ForwardingContent struct {
	FinalEndpoint   string          //the final endpoint to call
	DestinationNode *common.Address //null if its for everyone
	FinalParams     []byte          //the params to be passed at the end
	FinalReply      []byte          //ignored if DestinationNode is nill, otherwise will attempt to link back
	InitialSender   common.Address  //who started this
}

func (fc ForwardingContent) Hash() common.Hash {
	byteForm := []byte(fc.FinalEndpoint)
	if fc.DestinationNode != nil {
		byteForm = append(byteForm, fc.DestinationNode.Bytes()...)
	}
	byteForm = append(byteForm, fc.FinalParams...)
	byteForm = append(byteForm, fc.FinalReply...)

	return common.BytesToHash(crypto.Sha512(byteForm))
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
	ErrStateNotSet                = errors.New("StateDB was not established")
	ErrChainNotSet                = errors.New("chain reference not filled")
	ErrPreExistingAccount         = errors.New("specified account already exists on chain")
	ErrNoAccountSet               = errors.New("the account address has not been set")
	ErrNotSetupToHandleForwarding = errors.New("this RPC host is not setup to handle message forwarding")
	ErrAlreadyForwarded           = errors.New("message has already been forwarded")
	ErrBadForward                 = errors.New("this message has been deemed unfit to be shared further")
)
