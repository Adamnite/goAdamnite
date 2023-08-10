package blockchain

import (
	"sync"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/event"
	"github.com/adamnite/go-adamnite/params"
)

type TxPool struct {
	chainConfig *params.ChainConfig
	chain       blockChain
	mu          sync.RWMutex
	pending     map[common.Address]*txList
	all         *txLookup
	scope       event.SubscriptionScope
	locals      *accountSet
	txFeed      event.Feed
}

type blockChain interface {
	CurrentBlock() *types.Block
}

func (pool *TxPool) Get(txHash common.Hash) *types.Transaction {
	return pool.all.Get(txHash)
}

func (pool *TxPool) Pending() (map[common.Address]types.Transactions, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pending := make(map[common.Address]types.Transactions)
	for addr, list := range pool.pending {
		pending[addr] = list.Flatten()
	}
	return pending, nil
}

func (pool *TxPool) Locals() []common.Address {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.locals.flatten()
}

func (pool *TxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription {
	return pool.scope.Track(pool.txFeed.Subscribe(ch))
}
