package pendingHandling

import (
	"sort"

	"github.com/adamnite/go-adamnite/utils"
)

type TransactionQueue struct {
	pendingQueue   []*utils.Transaction
	pendingRemoval map[string]bool //uses the hash of the transaction to decide if it is set to be removed
}

func NewQueue() *TransactionQueue {
	tq := TransactionQueue{
		pendingRemoval: map[string]bool{},
	}
	return &tq
}
func (tq *TransactionQueue) AddToQueue(transaction *utils.Transaction) {
	if _, exists := tq.pendingRemoval[transaction.Hash().Hex()]; exists {
		//this already is awaiting processing, or has already been seen
		return
	}
	tq.pendingRemoval[transaction.Hash().Hex()] = false
	tq.pendingQueue = append(tq.pendingQueue, transaction)
}

// get the transaction that has been waiting in queue the longest
func (tq *TransactionQueue) Pop() *utils.Transaction {
	if len(tq.pendingQueue) == 0 {
		return nil
	}
	t := tq.pendingQueue[0]
	tq.pendingQueue = tq.pendingQueue[1:]
	if tq.pendingRemoval[t.Hash().Hex()] {
		return tq.Pop()
	}
	delete(tq.pendingRemoval, t.Hash().Hex())
	return t
}

// remove doesn't actually remove it from memory. But does make it so that it'll be removed next time there is a pop
func (tq *TransactionQueue) Remove(t *utils.Transaction) {
	tq.pendingRemoval[t.Hash().Hex()] = true
}

// removes all matching transactions from the queue
func (tq *TransactionQueue) RemoveAll(transactions []*utils.Transaction) {
	for _, t := range transactions {
		tq.pendingRemoval[t.Hash().Hex()] = true
	}
}

// sorts the queue so the transactions are returned oldest first!
func (tq *TransactionQueue) SortQueue() {
	sort.Slice(tq.pendingQueue, func(i, j int) bool {
		//TODO: move the ones awaiting removal to the front, then the oldest to newest
		if tq.pendingRemoval[tq.pendingQueue[i].Hash().Hex()] {
			//put the ones that need to be removed first
			return true
		}
		return tq.pendingQueue[i].Time.Before(tq.pendingQueue[j].Time)
	})
}
