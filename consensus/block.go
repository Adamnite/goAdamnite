package consensus

import (
	"crypto/sha256"
	"time"

	"github.com/adamnite/go-adamnite/common"
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
func NewBlock(parentBlockID common.Hash, witness common.Address, transactions []*Transaction) *Block {
	header := &BlockHeader{time.Now().Unix(), parentBlockID, witness, common.Hash{}, common.Hash{}, common.Hash{}}
	block := &Block{header, transactions}
	return block
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