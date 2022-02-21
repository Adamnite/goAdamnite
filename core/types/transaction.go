package types

import (
	"errors"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/adamnite/go-adamnite/common"
)

var (
	ErrorCryptoSignature    = errors.New("Invalid Elliptic Curve Values")
	ErrorEmptyTX            = errors.New("Empty Transaction Data")
	ErrorLowFee             = errors.New("Fee lower than minimum fee dicated by network")
	ErrorInvalidTransaction = errors.New("Transaction Balance is too large")
)

// Transaction is an Adamnite transaction.
type Transaction struct {
	InnerData Transaction_Data
	timestamp time.Time
	//Cache Values, similar to GoETH
	From atomic.Value
	hash atomic.Value
	size atomic.Value
}

func CreateTx(InnerData Transaction_Data) *Transaction {
	transaction := new(Transaction)
	transaction.Decode(InnerData.copy(), 0) //We will need to some functions for decoding and encoding data types
	//This can be RLP or our own implementation.
	return transaction
}

// TxData is the underlying data of a transaction.
type Transaction_Data interface {
	// copy creates a deep copy and initializes all fields
	copy() Transaction_Data
	chain_TYPE() *big.Int
	to() *common.Address
	amount() *big.Int
	message() []byte
	message_size() *big.Int //Size of message in bytes
	ATE_MAX() uint64        //Transaction Fee Limit that the user encoded
	ATE_price() *big.Int    //Actual Transaction Fee
	nonce() uint64

	rawSignature() (v, r, s *big.Int)
	setSignatue(chain_TYPE, v, r, s *big.Int)
}

type Transactions []*Transaction

//// TODO: Implement structure for creating a new transaction, basic message transactions (will do this weekend),
//proper encoding and decoding of transactions, fetch operations (the fetching of various transaction releated data),
//and some stack operations such as shift.

func (tx *Transaction) Decode(s Transaction_Data, _ int) {

}
