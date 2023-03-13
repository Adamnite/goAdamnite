// Package types contains the data structures that make up Adamnite's core protocol
package types

import (
	"math/big"
	"sync/atomic"
	"time"

	"github.com/adamnite/go-adamnite/common"
)

var (
	EmptyRootHash = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

type BlockHeader struct {
	ParentHash      common.Hash    `json:"parentHash" gencodec:"required"`  // The hash of the current block
	Time            uint64         `json:"timestamp" gencodec:"required"`   // The timestamp at which the block was approved
	Witness         common.Address `json:"witness" gencodec:"required"`     // The address of the witness that proposed the block
	WitnessRoot     common.Hash    `json:"witnessRoot" gencodec:"required"` // A hash of the witness state
	Number          *big.Int       `json:"number" gencodec:"required"`      // The block number of the current block
	Signature       common.Hash    `json:"signature" gencodec:"required"`   // The block signature that validates the block was created by right validator
	TransactionRoot common.Hash    `json:"txroot" gencodec:"required"`      // The root of the merkle tree in which transactions for this block are stored
	CurrentRound    uint64         `json:"round" gencodec:"required"`       // The current epoch number of the DPOS vote round
	StateRoot       common.Hash    `json:"stateRoot" gencodec:"required"`   // A hash of the current state
	Extra           []byte         `json:"extraData"        gencodec:"required"`
}

type Block struct {
	header          *BlockHeader
	transactionList Transactions

	//cache values
	hash atomic.Value
	size atomic.Value

	ReceivedAt   time.Time
	ReceivedFrom interface{}
}

type Body struct {
	Transactions []*Transaction
}

//// TODO: Implement structure for creating a new block, create a new block with
// header data, proper encoding of data, basic header checks (for example, check if the block number is too high),
// decoding of data, and functions to retrieve various header data and hashes.

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *BlockHeader) *BlockHeader {
	cpy := *h

	if cpy.Number = new(big.Int); h.Number != nil {
		cpy.Number.Set(h.Number)
	}

	return &cpy
}

func NewBlock(header *BlockHeader, txs []*Transaction, hasher TrieHasher) *Block {
	b := &Block{header: CopyHeader(header)}

	if len(txs) == 0 {
		b.header.TransactionRoot = EmptyRootHash
	} else {
		b.header.TransactionRoot = DeriveSha(Transactions(txs), hasher)
		b.transactionList = make(Transactions, len(txs))
		copy(b.transactionList, txs)
	}

	return b
}

func NewBlockWithHeader(header *BlockHeader) *Block {
	return &Block{header: CopyHeader(header)}
}

// Hash returns the block hash of the header, which is simply the keccak256 hash of its
// RLP encoding.
func (h *BlockHeader) Hash() common.Hash {
	return serializationHash(h)
}

func (b *Block) Hash() common.Hash {
	if hash := b.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := b.header.Hash()
	b.hash.Store(v)
	return v
}

func (b *Block) Number() *big.Int     { return new(big.Int).Set(b.header.Number) }
func (b *Block) Numberu64() uint64    { return b.header.Number.Uint64() }
func (b *Block) Body() *Body          { return &Body{b.transactionList} }
func (b *Block) Header() *BlockHeader { return CopyHeader(b.header) }
