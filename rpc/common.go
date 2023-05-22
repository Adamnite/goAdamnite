package rpc

import (
	"errors"
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/common"
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
	ErrStateNotSet                = errors.New("StateDB was not established")
	ErrChainNotSet                = errors.New("chain reference not filled")
	ErrPreExistingAccount         = errors.New("specified account already exists on chain")
	ErrNoAccountSet               = errors.New("the account address has not been set")
	ErrNotSetupToHandleForwarding = errors.New("this RPC host is not setup to handle message forwarding")
	ErrAlreadyForwarded           = errors.New("message has already been forwarded")
	ErrBadForward                 = errors.New("this message has been deemed unfit to be shared further")
)
