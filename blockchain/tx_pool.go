package blockchain

import (
	"math/big"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/utils/bytes"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/event"
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
	mu          sync.RWMutex
	pending     map[bytes.Address]*txList
	all         *txLookup
	scope       event.SubscriptionScope
	locals      *accountSet
	txFeed      event.Feed
}

type blockChain interface {
	CurrentBlock() *types.Block
	GetBlock(hash bytes.Hash, number *big.Int) *types.Block
	StateAt(root bytes.Hash) (*statedb.StateDB, error)
}

func NewTxPool(config TxPoolConfig, chainConfig *params.ChainConfig, chain blockChain) *TxPool {
	pool := &TxPool{
		config:      config,
		chainConfig: chainConfig,
		chain:       chain,

		pending: make(map[bytes.Address]*txList),
		all:     newTxLookup(),
	}

	return pool
}

func (pool *TxPool) Get(txHash bytes.Hash) *types.Transaction {
	return pool.all.Get(txHash)
}

func (pool *TxPool) Pending() (map[bytes.Address]types.Transactions, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pending := make(map[bytes.Address]types.Transactions)
	for addr, list := range pool.pending {
		pending[addr] = list.Flatten()
	}
	return pending, nil
}

func (pool *TxPool) Locals() []bytes.Address {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.locals.flatten()
}

func (pool *TxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription {
	return pool.scope.Track(pool.txFeed.Subscribe(ch))
}
