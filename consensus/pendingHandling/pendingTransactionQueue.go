package pendingHandling

import (
	"sort"
	"sync"

	"github.com/adamnite/go-adamnite/utils"
	"golang.org/x/sync/syncmap"
)

type TransactionQueue struct {
	pendingQueue   []*utils.Transaction
	pendingRemoval syncmap.Map
	//uses the hash of the transaction to decide if it is set to be removed
}

func NewQueue() *TransactionQueue {
	tq := TransactionQueue{
		pendingRemoval: sync.Map{},
	}
	return &tq
}
func (tq *TransactionQueue) AddToQueue(transaction *utils.Transaction) {

	if _, exists := tq.pendingRemoval.Load(transaction.Hash()); exists {
		//this already is awaiting processing, or has already been seen
		return
	}
	tq.pendingRemoval.Store(transaction.Hash(), false)
	tq.pendingQueue = append(tq.pendingQueue, transaction)
}

// get the transaction that has been waiting in queue the longest
func (tq *TransactionQueue) Pop() *utils.Transaction {
	if len(tq.pendingQueue) == 0 {
		return nil
	}
	t := tq.pendingQueue[0]
	tq.pendingQueue = tq.pendingQueue[1:]
	toDelete, previouslyStored := tq.pendingRemoval.LoadAndDelete(t.Hash())
	if !previouslyStored {
		//idk how we would have failed to store it in the map... but here we go returning it!
		return t
	}
	if toDelete.(bool) {
		return tq.Pop()
	}
	return t
}

// remove doesn't actually remove it from memory. But does make it so that it'll be removed next time there is a pop
func (tq *TransactionQueue) Remove(t *utils.Transaction) {
	tq.pendingRemoval.Store(t.Hash(), true)
}

// removes all matching transactions from the queue
func (tq *TransactionQueue) RemoveAll(transactions []*utils.Transaction) {
	for _, t := range transactions {
		tq.pendingRemoval.Store(t.Hash(), true)
	}
}

// sorts the queue so the transactions are returned oldest first!
func (tq *TransactionQueue) SortQueue() {
	sort.Slice(tq.pendingQueue, func(i, j int) bool {
		//TODO: move the ones awaiting removal to the front, then the oldest to newest
		toRemove, store := tq.pendingRemoval.Load(tq.pendingQueue[i].Hash())
		if !store || toRemove.(bool) {
			//put the ones that need to be removed(or arent stored) first
			return true
		}
		return tq.pendingQueue[i].Time.Before(tq.pendingQueue[j].Time)
	})
}
