package accounts

import (
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
	assert.Equal(t, 32, len(account.PrivateKey), "Private key should be 32 bytes long")
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
