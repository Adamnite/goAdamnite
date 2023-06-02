package consensus

import (
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/common"
)

type TransactionType 	    int8
type TransactionVersionType int8

const (
	TRANSACTION_EOA          TransactionType = iota // transaction between two EOA (Externally Owned Accounts)
	TRANSACTION_EOA_CONTRACT                        // transaction between EOA (Externally Owned Accounts) and contract
	TRANSACTION_CONTRACT                            // transaction between two contracts
	TRANSACTION_VOTE                                // transaction that doesn't transfer any coins but serves as a vote for a particular witness
)

const (
	// add more transaction versions as we develop our blockchain
	TRANSACTION_V1 TransactionVersionType = iota
)

type Transaction struct {
	Version TransactionVersionType
	Type    TransactionType

	Timestamp int64

	From *common.Address // sender's public address
	To   *common.Address // receiver's public address

	Amount   *big.Int // amount of NITE/Micalli from the sender to receiver
	GasPrice *big.Int // price of gas in Micalli, it's determined by market supply and demand
	GasLimit *big.Int // limit of the amount of NITE the sender is willing to pay for the transaction

	Data []byte // optional and used as input to the contracts
}

func NewTransaction(transactionType TransactionType, from *common.Address, to *common.Address, amount *big.Int, gasPrice *big.Int, gasLimit *big.Int) *Transaction {
	return &Transaction{
		TRANSACTION_V1,
		transactionType,
		time.Now().Unix(),
		from,
		to,
		amount,
		gasPrice,
		gasLimit,
		[]byte{},
	}
}

func NewTransactionWithData(transactionType TransactionType, from *common.Address, to *common.Address, amount *big.Int, gasPrice *big.Int, gasLimit *big.Int, data []byte) *Transaction {
	return &Transaction{
		TRANSACTION_V1,
		transactionType,
		time.Now().Unix(),
		from,
		to,
		amount,
		gasPrice,
		gasLimit,
		data,
	}
}
