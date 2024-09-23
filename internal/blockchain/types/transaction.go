package types

import "github.com/adamnite/go-adamnite/internal/common"

// Transaction types.
const (
	WitnessTxType = 0x00
	VoteTxType    = 0x01
	NormalTxType  = 0x02
)

type Transaction struct {
	txData TxData
}

type TxData interface {
	txType() byte
	copy() TxData

	getValidator() common.Address
}

// NewTx creates a new transaction.
func NewTx(inner TxData) *Transaction {
	tx := new(Transaction)
	tx.setDecoded(inner.copy(), 0)
	return tx
}

// setDecoded sets the inner transaction and size after decoding.
func (tx *Transaction) setDecoded(txData TxData, size uint64) {
	tx.txData = txData
}

func (tx *Transaction) Type() byte {
	return tx.txData.txType()
}

func (tx *Transaction) GetValidator() common.Address {
	return tx.txData.getValidator()
}
