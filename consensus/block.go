package consensus

import (
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils"
)

type Block struct {
	Header       *utils.BlockHeader
	Transactions []*utils.Transaction
	Signature    []byte
}

// NewBlock creates and returns Block
func NewBlock(parentBlockID common.Hash, witness crypto.PublicKey, witnessRoot common.Hash, transactionRoot common.Hash, stateRoot common.Hash, number *big.Int, transactions []*utils.Transaction) *Block {
	header := &utils.BlockHeader{
		Timestamp:             time.Now().UTC(),
		ParentBlockID:         parentBlockID,
		Witness:               witness,
		WitnessMerkleRoot:     witnessRoot,
		TransactionMerkleRoot: transactionRoot,
		StateMerkleRoot:       stateRoot,
		Number:                number,
	}
	block := &Block{
		Header:       header,
		Transactions: transactions,
	}
	return block
}

// ConvertHeader converts from the old block header structure to the new one (temporary workaround)
func ConvertBlockHeader(header *types.BlockHeader) *utils.BlockHeader {
	return &utils.BlockHeader{
		Timestamp:             time.Unix(int64(header.Time), 0),
		ParentBlockID:         header.ParentHash,
		Witness:               crypto.PublicKey(header.Witness[:]), //TODO: this wont actually work, but enough to compile
		WitnessMerkleRoot:     header.WitnessRoot,
		TransactionMerkleRoot: header.TransactionRoot,
		StateMerkleRoot:       header.StateRoot,
		Number:                header.Number,
	}
}

// ConvertBlock converts from the old block structure to the new one (temporary workaround)
func ConvertBlock(block *types.Block) *Block {
	return NewBlock(
		block.Header().ParentHash,
		crypto.PublicKey(block.Header().Witness[:]), //TODO: this wont actually work, but enough to compile
		block.Header().WitnessRoot,
		block.Header().TransactionRoot,
		block.Header().StateRoot,
		block.Number(),
		ConvertTransactions(block.Body().Transactions),
	)
}

// Hash gets block's header hash
func (b *Block) Hash() common.Hash {
	return b.Header.Hash()
}
