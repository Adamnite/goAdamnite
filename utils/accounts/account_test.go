package accounts

import (
	"bytes"
	"fmt"
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
		fmt.Println("msg not changes")
		t.Fail()
	}
	wrongAccount, _ := GenerateAccount()
	ans, err := wrongAccount.Decrypt(msg)
	if err == nil {
		fmt.Println("no error returned from wrong key")
		fmt.Println(ans)
		t.Fail()
	}

	ans, err = a.Decrypt(msg)
	if err != nil || !bytes.Equal(testMessageBytes, ans) {
		fmt.Println("bytes were not equal, or")
		t.Fatal(err)
	}
}
