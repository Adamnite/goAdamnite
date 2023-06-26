package utils

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
)

func TestTransaction(t *testing.T) {
	sender, _ := accounts.GenerateAccount()
	strippedSender := accounts.AccountFromPubBytes(sender.PublicKey)
	recipient, _ := accounts.GenerateAccount()

	tt, err := NewBaseTransaction(sender, recipient.Address, big.NewInt(5), big.NewInt(5))
	if err != nil {
		t.Fatal(err)
	}
	ok, err := tt.VerifySignature()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("signature of transaction could not be verified")
	}
	tt.From = strippedSender
	ok, err = tt.VerifySignature()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("signature of transaction could not be verified")
	}

}
func TestVMTransaction(t *testing.T) {
	sender, _ := accounts.GenerateAccount()
	strippedSender := accounts.AccountFromPubBytes(sender.PublicKey)
	recipient, _ := accounts.GenerateAccount()

	tt, err := NewBaseTransaction(sender, recipient.Address, big.NewInt(5), big.NewInt(5))
	if err != nil {
		t.Fatal(err)
	}
	vmt, err := NewVMTransactionFrom(sender, tt, []byte{1, 2, 3, 4, 5})
	if err != nil {
		t.Fatal(err)
	}

	ok, err := vmt.VerifySignature()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("signature of transaction could not be verified")
	}
	vmt.From = strippedSender
	ok, err = vmt.VerifySignature()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("signature of transaction could not be verified")
	}
	b, err := msgpack.Marshal(vmt)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	var nvmt *VMCallTransaction
	if err := msgpack.Unmarshal(b, &nvmt); err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t, vmt.VMInteractions.Hash(), nvmt.VMInteractions.Hash(),
		"not same after msgpack",
	)
	assert.Equal(
		t, vmt.Hash(), nvmt.Hash(),
		"hashes not equal after msgpack",
	)
	ok, err = nvmt.VerifySignature()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("signature of transaction could not be verified after marshal")
	}
}
