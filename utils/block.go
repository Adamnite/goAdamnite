package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
)

type BlockHeader struct {
	Timestamp time.Time // Timestamp at which the block was approved as Unix time stamp

	ParentBlockID common.Hash      // Hash of the parent block
	Witness       crypto.PublicKey // Address of the witness who proposed the block
	//TODO: change witnesses to be nodeID(pubkeys) in order!
	WitnessMerkleRoot     common.Hash // Merkle tree root in which witnesses for this block are stored
	TransactionMerkleRoot common.Hash // Merkle tree root in which transactions for this block are stored
	StateMerkleRoot       common.Hash // Merkle tree root in which states for this block are stored

	Number *big.Int
	Round  uint64
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
