package utils

import (
	"bytes"
	"encoding/binary"
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
	From           *accounts.Account
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
		From:     sender,
		To:       to,
		Time:     time.Now().UTC(),
		Amount:   value,
		GasLimit: gasLimit.Abs(gasLimit), //sanitizing the gas limit passed.
	}
	sig, err := sender.Sign(t)
	t.Signature = sig
	return &t, err
}
func (t Transaction) FromAddress() common.Address {
	return t.From.GetAddress()
}

var (
	ErrNegativeAmount  = fmt.Errorf("attempt to send negative funds")
	ErrIncorrectSigner = fmt.Errorf("attempt to sign from a different account")
)

// does not use the time in the hash value, as that is used for the random source value.
func (t Transaction) Hash() common.Hash {
	data := append(t.From.PublicKey, t.To.Bytes()...)
	data = append(data, t.Amount.Bytes()...)

	data = append(data, binary.LittleEndian.AppendUint64([]byte{}, uint64(t.Time.UnixMilli()))...)
	if t.VMInteractions != nil {
		data = append(data, t.VMInteractions.Hash().Bytes()...)
	}

	return common.BytesToHash(crypto.Sha512(data))
}

// Verify that the signature used in the transaction is correct
func (t Transaction) VerifySignature() (ok bool, err error) {
	if t.Amount.Sign() != 1 {
		return false, ErrNegativeAmount
	}
	return t.From.Verify(t, t.Signature), nil
}

// Test equality between two transactions
func (a Transaction) Equal(b Transaction) bool {
	return bytes.Equal(a.From.PublicKey, b.From.PublicKey) &&
		bytes.Equal(a.To.Bytes(), b.To.Bytes()) &&
		a.Amount.Cmp(b.Amount) == 0 &&
		bytes.Equal(a.Signature, b.Signature)
}
