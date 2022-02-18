package types

import (
	"math/big"

	"github.com/adamnite/go-adamnite/common"
)

type Blocks []*Block

type BlockHeader struct {
	ParentHash common.Hash    `json:"parentHash" gencodec:"required"`
	UncleHash  common.Hash    `json:"uncleHash" gencodec:"required"`
	Witness    common.Address `json:"validator" gencodec:"required"`
	TxHash     common.Hash    `json:"txHash" gencodec:"required"`
	Number     *big.Int       `json:"number" gencodec:"required"`
	GasLimit   uint64         `json:"gasLimit" gencodec:"required"`
	GasUsed    uint64         `json:"gasUsed" gencodec:"required"`
	Time       uint64         `json:"timestamp" gencodec:"required"`
}

type Block struct {
	header       *BlockHeader
	uncles       []*BlockHeader
	transactions Transactions
}
