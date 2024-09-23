package types

import (
	"math/big"

	"github.com/adamnite/go-adamnite/internal/common"
)

type Header struct {
	Number *big.Int
	Time   uint64
	Miner  common.Address
}

type Block struct {
	header Header
	txns   []Transaction
}

func NewBlock(header Header, txns []Transaction) *Block {
	block := &Block{
		header: header,
		txns:   txns,
	}

	return block
}

func (b *Block) GetBlockNumber() *big.Int {
	return b.header.Number
}

func (b *Block) GetTimestamp() uint64 {
	return b.header.Time
}

func (b *Block) GetTransactions() []Transaction {
	return b.txns
}
