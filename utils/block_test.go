package utils

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
	encoding "github.com/vmihailenco/msgpack/v5"
)

func TestBlocks(t *testing.T) {
	wit, _ := accounts.GenerateAccount()
	block := NewBlock(common.Hash{}, wit.PublicKey, common.Hash{}, common.Hash{}, common.Hash{}, big.NewInt(1), []*BaseTransaction{})
	hashA := block.Hash()
	if err := block.Sign(wit); err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t, hashA, block.Hash(),
		"hash did not return the same thing after being signed",
	)
	assert.True(t, block.VerifySignature(),
		"block signature not working",
	)
	b, err := encoding.Marshal(block)
	if err != nil {
		t.Fatal(err)
	}
	var block2 *Block
	if err := encoding.Unmarshal(b, &block2); err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t, block.Header.Hash(), block2.Header.Hash(),
		"headers not the same after msgpack",
	)
	assert.Equal(
		t, hashA.Bytes(), block2.Hash().Bytes(),
		"hash did not return the same thing after msgpack",
	)

	assert.True(t, block2.VerifySignature(),
		"block could not be verified after msgpack encoding",
	)
}
