package consensus

import (
	"crypto/sha256"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	encoding "github.com/vmihailenco/msgpack/v5"
)

type BlockHeader struct {
	Timestamp int64 // Timestamp at which the block was approved

	ParentBlockID common.Hash    // Hash of the parent block
	Witness       common.Address // Address of the witness who proposed the block

	WitnessMerkleRoot     common.Hash // Merkle tree root in which witnesses for this block are stored
	TransactionMerkleRoot common.Hash // Merkle tree root in which transactions for this block are stored
	StateMerkleRoot       common.Hash // Merkle tree root in which states for this block are stored
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
}

// NewBlock creates and returns Block
func NewBlock(parentBlockID common.Hash, witness common.Address, witnessRoot common.Hash, transactionRoot common.Hash, stateRoot common.Hash, transactions []*Transaction) *Block {
	header := &BlockHeader{time.Now().Unix(), parentBlockID, witness, witnessRoot, transactionRoot, stateRoot}
	block := &Block{header, transactions}
	return block
}

// ConvertBlock converts between old block structure and new one (temporary workaround)
func ConvertBlock(block *types.Block) *Block {
	return NewBlock(
		block.Header().ParentHash,
		block.Header().Witness,
		block.Header().WitnessRoot,
		block.Header().TransactionRoot,
		block.Header().StateRoot,
		ConvertTransactions(block.Body().Transactions),
	)
}

// Hash hashes block header
func (h *BlockHeader) Hash() common.Hash {
	bytes, _ := encoding.Marshal(h)

	sha := sha256.New()
	sha.Write(bytes)
	return common.BytesToHash(sha.Sum(nil))
}

// Hash gets block's header hash
func (b *Block) Hash() common.Hash {
	return b.Header.Hash()
}