package pendingHandling

import (
	"sync"

	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/safe"
	"golang.org/x/sync/syncmap"
)

type TransactionQueue struct {
	pendingQueue        *safe.SafeSlice
	pendingRemoval      syncmap.Map //uses the hash of the transaction to decide if it is set to be removed
	newTransactionsOnly bool        //if transactions that have already been reviewed be ignored when attempting to add them
	previouslySeen      syncmap.Map //hash of the transaction to if its new
	lock                sync.Mutex
}

func NewQueue(newOnly bool) *TransactionQueue {
	tq := TransactionQueue{
		pendingQueue:        safe.NewSafeSlice(),
		pendingRemoval:      sync.Map{},
		previouslySeen:      sync.Map{},
		newTransactionsOnly: newOnly,
		lock:                sync.Mutex{},
	}

	return &tq
}
func (tq *TransactionQueue) AddToQueue(transaction utils.TransactionType) {
	tq.lock.Lock()
	if tq.newTransactionsOnly {
		//this is to ignore anything that's already been seen
		_, previouslySeen := tq.previouslySeen.LoadOrStore(transaction, true)
		if previouslySeen {
			tq.lock.Unlock()
			return
		}
	}
	tq.lock.Unlock()
	tq.AddIgnoringPast(transaction)
}

// even if this has been reviewed, ignore that and add it
func (tq *TransactionQueue) AddIgnoringPast(transaction utils.TransactionType) {
	tq.lock.Lock()
	defer tq.lock.Unlock()
	if pendingRemoval, exists := tq.pendingRemoval.Load(transaction); exists && pendingRemoval.(bool) {
		//this already is awaiting processing, or has already been seen
		return
	}
	tq.pendingRemoval.Store(transaction, false)
	tq.pendingQueue.Append(transaction)
}

// get the transaction that has been waiting in queue the longest
func (tq *TransactionQueue) Pop() utils.TransactionType {
	tq.lock.Lock()
	if tq.pendingQueue.Len() == 0 {
		//if it's empty, we reset the pending removal as well
		tq.pendingRemoval.Range(func(key, _ any) bool {
			tq.pendingRemoval.Delete(key)
			return true
		})
		tq.lock.Unlock()
		return nil
	}
	tq.lock.Unlock()
	t := tq.pendingQueue.Pop(0)
	toDelete, exists := tq.pendingRemoval.LoadAndDelete(t)
	if exists && toDelete.(bool) {
		return tq.Pop()
	}
	if t == nil {
		return nil
	}
	return t.(utils.TransactionType)
}

// remove doesn't actually remove it from memory. But does make it so that it'll be removed next time there is a pop
func (tq *TransactionQueue) Remove(t utils.TransactionType) {
	tq.lock.Lock()
	defer tq.lock.Unlock()
	tq.pendingRemoval.Store(t, true)
	tq.previouslySeen.Store(t, true)
}

// removes all matching transactions from the queue. Only works with direct interface
func (tq *TransactionQueue) RemoveAll(transactions []utils.TransactionType) {
	tq.lock.Lock()
	defer tq.lock.Unlock()
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
	tq.lock.Lock()
	defer tq.lock.Unlock()
	tq.pendingQueue.Sort(func(a, b any) bool {
		i := a.(utils.TransactionType)
		j := b.(utils.TransactionType)
		toRemove, store := tq.pendingRemoval.Load(i)
		if !store || toRemove.(bool) {
			//put the ones that need to be removed(or aren't stored) first
			return true
		}
		return i.GetTime().Before(j.GetTime())
	})
}
