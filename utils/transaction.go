package utils

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
)

type Transaction struct {
	From      common.Address
	To        common.Address
	Amount    *big.Int
	Time      time.Time
	Signature []byte
}

var (
	ErrNegativeAmount  = fmt.Errorf("attempt to send negative funds")
	ErrIncorrectSigner = fmt.Errorf("attempt to sign from a different account")
)

// signs the transaction. Resets the time to now after signing.
func (t *Transaction) Sign(key ecdsa.PrivateKey) (err error) { //TODO: replace with our key library
	hash := t.Hash()
	t.Signature, err = key.Sign(rand.New(rand.NewSource(t.Time.Unix())), hash.Bytes(), nil)
	t.Time = time.Now()
	return err
}

// does not use the time in the hash value, as that is used for the random source value.
func (t *Transaction) Hash() common.Hash {
	data := append(t.From.Bytes(), t.To.Bytes()...)
	data = append(data, t.Amount.Bytes()...)

	return common.BytesToHash(crypto.Sha512(data))
}

// Verify that the signature used in the transaction is correct
func (t *Transaction) VerifySignature(key ecdsa.PublicKey) (ok bool, err error) {
	if t.Amount.Sign() != 1 {
		return false, ErrNegativeAmount
	}
	if t.From != crypto.PubkeyToAddress(key) {
		return false, ErrIncorrectSigner
	}
	ok = ecdsa.VerifyASN1(&key, t.Hash().Bytes(), t.Signature)
	return ok, nil
}

// Test equality between two transactions
func (a Transaction) Equal(b Transaction) bool {
	return bytes.Equal(a.From.Bytes(), b.From.Bytes()) &&
		bytes.Equal(a.To.Bytes(), b.To.Bytes()) &&
		a.Amount.Cmp(b.Amount) == 0 &&
		bytes.Equal(a.Signature, b.Signature)
}
