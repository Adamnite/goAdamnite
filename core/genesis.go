package core

import (
	"math/big"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/params"
)

type Genesis struct {
	Config          *params.ChainConfig `json:"config"`
	Timestamp       uint64              `json:"timestamp"`
	GasLimit        uint64              `json:"gasLimit" gencodec:"required"`
	Coinbase        common.Address      `json:"address"`
	WitnessListHash common.Hash         `json:"witnessListHash"`
	Alloc           GenesisAlloc        `json:"alloc" gencodec:"required"`

	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`
}

type GenesisAlloc map[common.Address]GenesisAccount

type GenesisAccount struct {
	PrivateKey []byte
	Balance    *big.Int
	Nonce      uint64
}
