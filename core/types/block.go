//Package Main contains the data structures that make up Adamnite's core protocol
package main

import (
	"math/big"
  "sync/atomic"
	"time"
	"github.com/adamnite/go-adamnite/common"
)


type BlockHeader struct {
	PreviousHash common.Hash  `json:"PreviousHash" gencodec:"required"`
  Hash       common.Hash    `json:"Hash" gencodec:"required"`
  Time       uint64         `json:"timestamp" gencodec:"required"`
	Witness    common.Address `json:"witness" gencodec:"required"`
  WitnessListHash     common.Hash       `json:"witnessListHash" gencodec:"required"`
  NetFee     uint64         `json:"NetFee" gencodec:"required"`
  NetStorage uint64         `json:"NetStorage" gencodec:"required"`
	Number     *big.Int       `json:"number" gencodec:"required"`
  Signature  common.Hash    `json:"number" gencodec:"required"`
  Size       uint64         `json:"Size" gencodec:"required"`
	TransactionRoot   common.Hash        `json:"TransactionRoot" gencodec:"required"`
}

type Block struct {
	header       *BlockHeader
	transactionList Transactions
  witnessList []*BlockHeader
  //cache values, similar to GoETH
  hash atomic.Value
	size atomic.Value

}

//// TODO: Implement structure for creating a new block, create a new block with
//header data, proper encoding of data, basic header checks (for example, check if the block number is too high),
// decoding of data, and functions to retreive various header data and hashes.
