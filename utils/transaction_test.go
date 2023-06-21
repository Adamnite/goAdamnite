package utils

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/utils/accounts"
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
