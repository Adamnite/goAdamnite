//Package types contains the data structures that make up Adamnite's core protocol
package types

import (
	"io"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	EmptyRootHash = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

type BlockHeader struct {
	PreviousHash	common.Hash			`json:"PreviousHash" gencodec:"required"`
	Witness			common.Address		`json:"Witness" gencodec:"required"`
	NetFee			*big.Int	        `json:"NetFee" gencodec:"required"`
	Nonce			*big.Int		`json:"Nonce" gencodec:"required"`
	TransactionRoot	common.Hash		        `json:"TransactionRoot" gencodec:"required"`
	Signature		[8]byte			`json:"Signature" gencodec:"required"`
	Timestamp		uint64			`json:"Timestamp" gencodec:"required"`
	Extra			[]byte		        `json:"Extra" gencodec:"required"`
}

type Block struct {
	header          *BlockHeader
	transaction_list Transactions

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
// decoding of data, and functions to retreive various header data and hashes.

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

// Hash returns the block hash of the header, which is simply the keccak256 hash of its
// Serialization encoding.
func (h *BlockHeader) Hash() common.Hash {
	return SerializationHash(h)
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

type encodingBlock struct {
	Header *BlockHeader
	Txs    []*Transaction
}

func (b *Block) EncodeSerialization(w io.Writer) error {
	return msgpack.NewEncoder(w).Encode(encodingBlock{
		Header: b.header,
		Txs:    b.transactionList,
	})
}

func (b *Block) DecodeSerialization(s *msgpack.Decoder) error {

	var eb encodingBlock
	size, _ := s.DecodeBytesLen()
	if err := s.Decode(&eb); err != nil {
		return err
	}
	b.header, b.transactionList = eb.Header, eb.Txs
	b.size.Store(common.StorageSize(size))
	return nil
}
