package pendingHandling

import (
	"sort"
	"sync"

	"github.com/adamnite/go-adamnite/utils"
	"golang.org/x/sync/syncmap"
)

type TransactionQueue struct {
	pendingQueue        []utils.TransactionType
	pendingRemoval      syncmap.Map //uses the hash of the transaction to decide if it is set to be removed
	newTransactionsOnly bool        //if transactions that have already been reviewed be ignored when attempting to add them
	previouslySeen      syncmap.Map //hash of the transaction to if its new
}

func NewQueue(newOnly bool) *TransactionQueue {
	tq := TransactionQueue{
		pendingRemoval:      sync.Map{},
		previouslySeen:      sync.Map{},
		newTransactionsOnly: newOnly,
	}
	return &tq
}
func (tq *TransactionQueue) AddToQueue(transaction utils.TransactionType) {
	if tq.newTransactionsOnly {
		//this is to ignore anything that's already been seen
		_, previouslySeen := tq.previouslySeen.LoadOrStore(transaction, true)
		if previouslySeen {
			return
		}
	}
	tq.AddIgnoringPast(transaction)
}

// even if this has been reviewed, ignore that and add it
func (tq *TransactionQueue) AddIgnoringPast(transaction utils.TransactionType) {
	if pendingRemoval, exists := tq.pendingRemoval.Load(transaction); exists && pendingRemoval.(bool) {
		//this already is awaiting processing, or has already been seen
		return
	}
	tq.pendingRemoval.Store(transaction, false)
	tq.pendingQueue = append(tq.pendingQueue, transaction)
}

// get the transaction that has been waiting in queue the longest
func (tq *TransactionQueue) Pop() utils.TransactionType {
	if len(tq.pendingQueue) == 0 {
		//if it's empty, we reset the pending removal as well
		tq.pendingRemoval = sync.Map{}
		return nil
	}
	t := tq.pendingQueue[0]
	tq.pendingQueue = tq.pendingQueue[1:]
	toDelete, exists := tq.pendingRemoval.LoadAndDelete(t)
	if exists && toDelete.(bool) {
		return tq.Pop()
	}
	return t
}

// remove doesn't actually remove it from memory. But does make it so that it'll be removed next time there is a pop
func (tq *TransactionQueue) Remove(t utils.TransactionType) {
	tq.pendingRemoval.Store(t, true)
	tq.previouslySeen.Store(t, true)
}

// removes all matching transactions from the queue. Only works with direct interface
func (tq *TransactionQueue) RemoveAll(transactions []utils.TransactionType) {
	for _, t := range transactions {
		tq.pendingRemoval.Store(t, true)
		tq.previouslySeen.Store(t, true)
	}
}

// same as removeAll but works with any transaction type
func RemoveAllFrom[T utils.TransactionType](transactions []T, tq *TransactionQueue) {
	for _, t := range transactions {
		tq.Remove(t)
	}
}

// sorts the queue so the transactions are returned oldest first!
func (tq *TransactionQueue) SortQueue() {
	sort.Slice(tq.pendingQueue, func(i, j int) bool {
		//TODO: move the ones awaiting removal to the front, then the oldest to newest
		toRemove, store := tq.pendingRemoval.Load(tq.pendingQueue[i])
		if !store || toRemove.(bool) {
			//put the ones that need to be removed(or aren't stored) first
			return true
		}
		return tq.pendingQueue[i].GetTime().Before(tq.pendingQueue[j].GetTime())
	})
}
