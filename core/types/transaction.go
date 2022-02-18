package types

import (
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/common"
)

// Transactions implements DerivableList for transactions.
type Transactions []*Transaction

// Transaction is an Adamnite transaction.
type Transaction struct {
	txdata TxData
	time   time.Time
}

// TxData is the underlying data of a transaction.
type TxData interface {
	// txType returns the type ID
	txType() byte

	// copy creates a deep copy and initializes all fields
	copy() TxData

	chainID() *big.Int
	data() []byte
	gas() uint64
	gasPrice() *big.Int
	value() *big.Int
	nonce() uint64
	to() *common.Address

	rawSignatureValues() (v, r, s *big.Int)
	setSignatureValues(chainID, v, r, s *big.Int)
}
