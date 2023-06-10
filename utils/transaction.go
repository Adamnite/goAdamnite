package utils

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

type TransactionVersionType int8

const (
	// add more transaction versions as we develop our blockchain
	TRANSACTION_V0 TransactionVersionType = iota
)

type Transaction struct {
	Version        TransactionVersionType
	From           common.Address
	To             common.Address
	Amount         *big.Int
	GasLimit       *big.Int
	Time           time.Time
	VMInteractions *RuntimeChanges //an optional value that would make this transaction require chamber B validation
	Signature      []byte
}

func NewTransaction(sender *accounts.Account, to common.Address, value *big.Int, gasLimit *big.Int) (*Transaction, error) {
	if value.Sign() == -1 {
		return nil, ErrNegativeAmount
	}
	t := Transaction{
		Version:  TRANSACTION_V0,
		From:     sender.Address,
		To:       to,
		Amount:   value,
		GasLimit: gasLimit.Abs(gasLimit), //sanitizing the gas limit passed.
	}
	sig, err := sender.Sign(t)
	t.Signature = sig
	return &t, err
}

var (
	ErrNegativeAmount  = fmt.Errorf("attempt to send negative funds")
	ErrIncorrectSigner = fmt.Errorf("attempt to sign from a different account")
)

// does not use the time in the hash value, as that is used for the random source value.
func (t *Transaction) Hash() common.Hash {
	data := append(t.From.Bytes(), t.To.Bytes()...)
	data = append(data, t.Amount.Bytes()...)
	timeBytes, _ := t.Time.MarshalBinary()
	data = append(data, timeBytes...)
	if t.VMInteractions != nil {
		data = append(data, t.VMInteractions.Hash().Bytes()...)
	}

	return common.BytesToHash(crypto.Sha512(data))
}

// Verify that the signature used in the transaction is correct
func (t *Transaction) VerifySignature(signer accounts.Account) (ok bool, err error) {
	if t.Amount.Sign() != 1 {
		return false, ErrNegativeAmount
	}
	return signer.Verify(t, t.Signature), nil
}

// Test equality between two transactions
func (a Transaction) Equal(b Transaction) bool {
	return bytes.Equal(a.From.Bytes(), b.From.Bytes()) &&
		bytes.Equal(a.To.Bytes(), b.To.Bytes()) &&
		a.Amount.Cmp(b.Amount) == 0 &&
		bytes.Equal(a.Signature, b.Signature)
}
