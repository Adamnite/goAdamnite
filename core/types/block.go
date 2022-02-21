//Package types contains the data structures that make up Adamnite's core protocol
package types

import (
	"math/big"
	"sync/atomic"

	"github.com/adamnite/go-adamnite/common"
)

type BlockHeader struct {
	PreviousHash    common.Hash    `json:"PreviousHash" gencodec:"required"`    // The hash of the previous block
	Hash            common.Hash    `json:"Hash" gencodec:"required"`            // The hash of the current block
	Time            uint64         `json:"timestamp" gencodec:"required"`       // The timestamp at which the block was approved
	Witness         common.Address `json:"witness" gencodec:"required"`         // The address of the witness that proposed the block
	WitnessListHash common.Hash    `json:"witnessListHash" gencodec:"required"` // A hash of the witness list
	NetFee          uint64         `json:"NetFee" gencodec:"required"`          // The total amount of transaction fees for the current block
	NetStorage      uint64         `json:"NetStorage" gencodec:"required"`      // The net size, in bytes, of all the messages in the transactions for this block
	Number          *big.Int       `json:"number" gencodec:"required"`          // The block number of the current block
	Signature       common.Hash    `json:"number" gencodec:"required"`          // The block signature that validates the block was created by right validator
	Size            uint64         `json:"Size" gencodec:"required"`            // The overall size in bytes of the block
	TransactionRoot common.Hash    `json:"TransactionRoot" gencodec:"required"` // The root of the merkle tree in which transactions for this block are stored
}

type Block struct {
	header          *BlockHeader
	transactionList Transactions
	witnessList     []*BlockHeader
	//cache values, similar to GoETH
	hash atomic.Value
	size atomic.Value
}

//// TODO: Implement structure for creating a new block, create a new block with
// header data, proper encoding of data, basic header checks (for example, check if the block number is too high),
// decoding of data, and functions to retreive various header data and hashes.
