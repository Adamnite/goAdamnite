package tests

import (
	"testing"

	"github.com/adamnite/go-adamnite/crypto"
)

func TestCreateAddressTest(t *testing.T) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Error("Error in crypto generatekey function")
	}

	pubkey := privKey.PublicKey
	address := crypto.PubkeyToAddress(pubkey)
	t.Log(address.String())
}

func TestSignAndValidate(t *testing.T) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Error("Error in GenerateKey function")
	}

	var datahash = []byte("abcdefghijklmnopqrstuvwxyz012345")
	signature, err := crypto.Sign(datahash, privKey)
	if err != nil {
		t.Error("Error in sign function")
	}

	ret := crypto.VerifySignature(crypto.FromECDSAPub(&privKey.PublicKey), datahash, signature[0:64])

	if ret == false {
		t.Error("validate error")
	}
}
