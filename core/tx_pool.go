package core

import (
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/params"
)

type TxPoolConfig struct {
	Lifetime time.Duration // Maximum amount of time non-executable transaction are queued.

	AccountSlots uint64
	AccountQueue uint64
	GlobalSlots  uint64
	GlobalQueue  uint64
}

var DefaultTxPoolConfig = TxPoolConfig{
	Lifetime: 1 * time.Hour,

	AccountSlots: 32,
	AccountQueue: 64,
	GlobalSlots:  8192,
	GlobalQueue:  1024,
}

type TxPool struct {
	config      TxPoolConfig
	chainConfig *params.ChainConfig
	chain       blockChain

	pending map[common.Address]*txList
}

type blockChain interface {
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*statedb.StateDB, error)
}

func NewTxPool(config TxPoolConfig, chainConfig *params.ChainConfig, chain blockChain) *TxPool {
	pool := &TxPool{
		config:      config,
		chainConfig: chainConfig,
		chain:       chain,

		pending: make(map[common.Address]*txList),
	}

	return pool
}
