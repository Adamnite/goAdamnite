package accounts

import (
	"bytes"
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAccount(t *testing.T) {
	account, err := GenerateAccount()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 28, len(account.Address), "Address should be 28 bytes long")
	assert.Equal(t, 65, len(account.PublicKey), "Public key should be 65 bytes long")
	assert.Equal(t, 32, len(account.privateKey), "Private key should be 32 bytes long")
}

func TestSignData(t *testing.T) {
	account, err := GenerateAccount()
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("Test message")
	signature, err := account.Sign(data)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 65, len(signature), "Signature should be 65 bytes long")
	assert.True(t, account.Verify(data, signature), "Signature should be verified")
}

func TestMultipleSignTypes(t *testing.T) {
	bData := []byte("Test message")
	signAndVerify(t, bData)
	signAndVerify(t, "test string")
	signAndVerify(t, big.NewInt(1))
}

func signAndVerify(t *testing.T, val interface{}) {
	account, err := GenerateAccount()
	if err != nil {
		t.Fatal(err)
	}
	if signatures, err := account.Sign(val); err != nil {
		t.Fatal(err)
	} else {
		assert.True(
			t,
			account.Verify(val, signatures),
			"bytes mismatch to signature",
		)
	}
}

func TestEncryption(t *testing.T) {
	a, _ := GenerateAccount()
	senderAccount := AccountFromPubBytes(a.PublicKey)
	testMessageBytes := []byte("Hello World!")
	msg, err := senderAccount.Encrypt(testMessageBytes)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(testMessageBytes, msg) {
		fmt.Println("msg not changed")
		t.Fail()
	}
	uninvolvedAccount, _ := GenerateAccount()
	ans, err := uninvolvedAccount.Decrypt(msg)
	if err == nil || bytes.Equal(ans, testMessageBytes) {
		fmt.Println("successful message decryption with another account")
		fmt.Println(ans)
		t.Fail()
	}

	ans, err = a.Decrypt(msg)
	if err != nil || !bytes.Equal(testMessageBytes, ans) {
		fmt.Println("bytes were not equal, or")
		t.Fatal(err)
	}
}

func TestLotsOfAccounts(t *testing.T) {
	numberOfAccounts := 10000
	testData := "Hello World!"
	accounts := make([]*Account, numberOfAccounts)
	pubOnlyAccounts := make([]*Account, numberOfAccounts)
	for i := 0; i < numberOfAccounts; i++ {
		ac, err := GenerateAccount()
		if err != nil {
			t.Fatal(err)
		}
		accounts[i] = ac
		pubOnlyAccounts[i] = AccountFromPubBytes(ac.PublicKey)
	}

	//the accounts are all made. If any of them are gonna fail, we want them to fail before this point!

	for i := 0; i < numberOfAccounts; i++ {
		signature, err := accounts[i].Sign(testData)
		if err != nil {
			log.Printf("privatekey that was found invalid is %X", accounts[i].privateKey)
			if l := len(accounts[i].privateKey); l != 32 {
				log.Printf("privatekey length is %v, which is invalid", l)
			}
			t.Fatal(err)
		}
		assert.True(
			t,
			accounts[i].Verify(testData, signature),
			"could not verify with full account",
		)
		assert.True(
			t,
			pubOnlyAccounts[i].Verify(testData, signature),
			"could not verify with pubKey only account",
		)
		encrypted, err := pubOnlyAccounts[i].Encrypt([]byte(testData))
		if err != nil {
			t.Fatal(err)
		}
		_, err = pubOnlyAccounts[i].Decrypt(encrypted)
		//should throw an error
		if err == nil {
			t.Error("pubOnly account did not throw an error when attempting to decrypt a message")
		}
		ans, err := accounts[i].Decrypt(encrypted)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(
			t,
			[]byte(testData),
			ans,
			"decrypted message at large scale was not same as before encryption",
		)
	}
}
