package rpc

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
	encoding "github.com/vmihailenco/msgpack/v5"
)

//for testing bouncer, since bouncer endpoints don't have their own client (necessarily)

func TestGetBalance(t *testing.T) {
	for i, ac := range testAccounts {
		ansBytes := []byte{}
		accountString := ac.Hex()
		addressBytes, _ := encoding.Marshal(&accountString)
		if err := bouncerClient.Call(getBalanceEndpoint, addressBytes, &ansBytes); err != nil {
			t.Fatal(err)
		}
		ansString := ""
		if err := encoding.Unmarshal(ansBytes, &ansString); err != nil {
			t.Fatal(err)
		}
		balance, ok := big.NewInt(0).SetString(ansString, 0)
		assert.True(t, ok, "balance could not be recovered")
		assert.Equal(t, testBalances[i], balance, "balances are not equal")
	}
}

func TestGetAddresses(t *testing.T) {
	for _, ac := range testAccounts {
		accountString := ac.Hex()
		addressBytes, _ := encoding.Marshal(&accountString)
		ansBytes := []byte{}
		if err := bouncerClient.Call(createAccountEndpoint, addressBytes, &ansBytes); err != nil {
			t.Fatal(err)
		}
		var ok bool
		if err := encoding.Unmarshal(ansBytes, &ok); err != nil {
			t.Fatal(err)
		}
		assert.True(t, ok, "account could not be added")
	}
	//check that it doesn't allow duplicate account creation
	for _, ac := range testAccounts {
		accountString := ac.Hex()
		addressBytes, _ := encoding.Marshal(&accountString)
		ansBytes := []byte{}
		if err := bouncerClient.Call(createAccountEndpoint, addressBytes, &ansBytes); err.Error() != ErrPreExistingAccount.Error() {
			t.Fatal(err)
		}

	}
}
func TestMessaging(t *testing.T) {
	sender, _ := accounts.GenerateAccount()
	recipient, _ := accounts.GenerateAccount()
	msg, err := utils.NewCaesarMessage(*recipient, *sender, "Hello World!")
	if !msg.Verify() {
		panic("ahh!")
	}
	if err != nil {
		t.Fatal(err)
	}
	msgBytes, _ := encoding.Marshal(msg)
	err = bouncerClient.Call(bouncerNewMessageEndpoint, msgBytes, []byte{}) //this will return an error (it doesn't have a forwarding server of its own)
	assert.Equal(t,
		"this is an incomplete bouncer server, and cannot forward",
		err.Error(), "different error returned? %v", err)

}
