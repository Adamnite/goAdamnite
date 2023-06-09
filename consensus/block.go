package consensus

import (
	"crypto/sha256"
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	encoding "github.com/vmihailenco/msgpack/v5"
)

type BlockHeader struct {
	Timestamp int64 // Timestamp at which the block was approved as Unix time stamp

	ParentBlockID common.Hash    // Hash of the parent block
	Witness       common.Address // Address of the witness who proposed the block
	//TODO: change witnesses to be nodeID(pubkeys) in order!
	WitnessMerkleRoot     common.Hash // Merkle tree root in which witnesses for this block are stored
	TransactionMerkleRoot common.Hash // Merkle tree root in which transactions for this block are stored
	StateMerkleRoot       common.Hash // Merkle tree root in which states for this block are stored

	Number *big.Int
	Round  uint64
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	Signature    []byte
}

// NewBlock creates and returns Block
func NewBlock(parentBlockID common.Hash, witness common.Address, witnessRoot common.Hash, transactionRoot common.Hash, stateRoot common.Hash, number *big.Int, transactions []*Transaction) *Block {
	header := &BlockHeader{
		Timestamp:             time.Now().Unix(),
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
func ConvertBlockHeader(header *types.BlockHeader) *BlockHeader {
	return &BlockHeader{
		Timestamp:             int64(header.Time),
		ParentBlockID:         header.ParentHash,
		Witness:               header.Witness,
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
		block.Header().Witness,
		block.Header().WitnessRoot,
		block.Header().TransactionRoot,
		block.Header().StateRoot,
		block.Number(),
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
