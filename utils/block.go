package utils

//TODO: once this block is the only version, it (as well as transactions, and similar files) should be moved to RPC
import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

type BlockHeader struct {
	Timestamp time.Time // Timestamp at which the block was approved as Unix time stamp

	ParentBlockID         common.Hash      // Hash of the parent block
	Witness               crypto.PublicKey // Address of the witness who proposed the block
	WitnessMerkleRoot     common.Hash      // Merkle tree root in which witnesses for this block are stored
	TransactionMerkleRoot common.Hash      // Merkle tree root in which transactions for this block are stored
	StateMerkleRoot       common.Hash      // Merkle tree root in which states for this block are stored
	TransactionType       int8
	Number                *big.Int //the block number
	Round                 uint64
}

// Hash hashes block header
func (h *BlockHeader) Hash() common.Hash {
	val, _ := h.Timestamp.MarshalBinary()
	val = append(val, h.ParentBlockID[:]...)
	val = append(val, h.Witness[:]...)
	val = append(val, h.WitnessMerkleRoot[:]...)
	val = append(val, h.StateMerkleRoot[:]...)
	val = append(val, h.TransactionMerkleRoot[:]...)
	val = append(val, h.Number.Bytes()...)
	val = binary.BigEndian.AppendUint64(val, h.Round)

	sha := sha256.New()
	sha.Write(val)
	return common.BytesToHash(sha.Sum(nil))
}

type Block struct {
	Header       *BlockHeader
	Transactions []TransactionType
	Signature    []byte
}

// NewBlock creates and returns Block
func NewBlock(parentBlockID common.Hash, witness crypto.PublicKey, witnessRoot common.Hash, transactionRoot common.Hash, stateRoot common.Hash, number *big.Int, transactions []TransactionType) *Block {
	header := &BlockHeader{
		Timestamp:             time.Now().UTC(),
		ParentBlockID:         parentBlockID,
		Witness:               witness,
		WitnessMerkleRoot:     witnessRoot,
		TransactionMerkleRoot: transactionRoot,
		StateMerkleRoot:       stateRoot,
		Number:                number,
	}
	header.TransactionType = 0x02 //TODO: make this correct
	block := &Block{
		Header:       header,
		Transactions: transactions,
	}
	return block
}

// Hash gets block's hash
func (b *Block) Hash() common.Hash {
	bytes := b.Header.Hash().Bytes()
	for _, t := range b.Transactions {
		bytes = append(bytes, (t).GetSignature()...)
		//since the signature normally isn't added to the transactions hash (how would you sign the hash of a signature of a hash of a....)
	}

	return common.BytesToHash(crypto.Sha512(bytes))
}

func (b *Block) Sign(signer accounts.Account) error {
	sig, err := signer.Sign(b)
	b.Signature = sig
	return err
}
func (b *Block) VerifySignature() bool {
	if b.Header == nil || b.Header.Witness == nil {
		return false //might as well add safe guards
	}
	signer := accounts.AccountFromPubBytes(b.Header.Witness)
	return signer.Verify(b, b.Signature)
}
