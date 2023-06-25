package rpc

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
	encoding "github.com/vmihailenco/msgpack/v5"
)

// for testing bouncer, since bouncer endpoints don't have their own client (necessarily)

func TestGetChainID(t *testing.T) {
	output := []byte{}

	err := bouncerClient.Call(getChainIDEndpoint, []byte{}, &output)
	if err != nil {
		t.Fatal(err)
	}

	var version string
	if err := encoding.Unmarshal(output, &version); err != nil {
		t.Fatal(err)
	}

	if !assert.Equal(t, "0.1.2", version, "Chain ID is not correct") {
		t.Fail()
	}
}

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
	for _, acc := range testAccounts {
		account := struct {
			Address string
		}{ acc.Hex() }

		accountData, err := encoding.Marshal(&account)
		if err != nil {
			t.Fatal(err)
		}

		output := []byte{}
		if err := bouncerClient.Call(createAccountEndpoint, accountData, &output); err != nil {
			t.Fatal(err)
		}

		var ok bool
		if err := encoding.Unmarshal(output, &ok); err != nil {
			t.Fatal(err)
		}
		assert.True(t, ok, "account could not be added")
	}

	// check that duplicate account creation is not allowed
	for _, acc := range testAccounts {
		account := struct {
			Address string
		}{ acc.Hex() }

		accountData, err := encoding.Marshal(&account)
		if err != nil {
			t.Fatal(err)
		}

		output := []byte{}
		if err := bouncerClient.Call(createAccountEndpoint, accountData, &output); err.Error() != ErrPreExistingAccount.Error() {
			t.Fatal(err)
		}
	}
}

func TestMessaging(t *testing.T) {
	sender, _ := accounts.GenerateAccount()
	receiver, _ := accounts.GenerateAccount()

	message := []byte("Hello, world!")
	encryptedMessage, err := receiver.Encrypt(message)
	if err != nil {
		t.Fatal()
	}

	signedMsg, err := sender.Sign(encryptedMessage)
	if err != nil {
		t.Fatal(err)
	}

	input := struct {
		FromPublicKey string
		ToPublicKey   string
		RawMessage    string
		SignedMessage string
	}{
		common.ToHex(sender.PublicKey),
		common.ToHex(receiver.PublicKey),
		common.ToHex(encryptedMessage),
		common.ToHex(signedMsg),
	}

	msgData, err := encoding.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}

	output := []byte{}
	err = bouncerClient.Call(bouncerNewMessageEndpoint, msgData, &output) //this will return an error (it doesn't have a forwarding server of its own)
	assert.Equal(
		t,
		"this is an incomplete bouncer server, and cannot forward",
		err.Error(), "different error returned? %v", err)
}
