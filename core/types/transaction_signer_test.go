package types

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
)

func TestAdamniteSigner(t *testing.T) {
	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)

	signer := AdamniteSigner{}

	tx, err := SignTransaction(NewVoteTransaction(0, common.HexToAddress("0x2d9487a9551db05414018c7fac9aed393f2fccda"), new(big.Int), 0), signer, key)
	if err != nil {
		t.Fatal(err)
	}

	from, err := Sender(signer, tx)
	if err != nil {
		t.Fatal(err)
	}

	if from != addr {
		t.Errorf("exected from and addr to be equal. Got %x want %x", from, addr)
	}
}
