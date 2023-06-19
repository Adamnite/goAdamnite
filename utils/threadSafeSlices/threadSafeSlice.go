package threadSafeSlice

import (
	"fmt"
	"sync"
)

type ThreadSafeSlice struct {
	lock  sync.RWMutex
	items []interface{}
}

type ThreadSafeItem struct {
	Index int
	Value interface{}
}

func NewThreadSafeSlice() *ThreadSafeSlice {
	return &ThreadSafeSlice{}
}

func (tss *ThreadSafeSlice) Copy() *ThreadSafeSlice {
	tss.lock.RLock()
	defer tss.lock.RUnlock()
	newTSS := ThreadSafeSlice{
		items: make([]interface{}, len(tss.items)),
		lock:  sync.RWMutex{},
	}
	copy(newTSS.items, tss.items)
	return &newTSS
}

// removes an item from the array based on it's index
func (tss *ThreadSafeSlice) Remove(index int) {
	tss.lock.Lock()
	defer tss.lock.Unlock()
	if index < 0 {
		index = len(tss.items) + index
	}
	tss.items = append(tss.items[:index], tss.items[index+1:]...)
}

// removes all the items from a-b including a and b
func (tss *ThreadSafeSlice) RemoveFrom(a, b int) {
	tss.lock.Lock()
	defer tss.lock.Unlock()
	itemsLength := len(tss.items)
	var aIndex int = a
	if a < 0 {
		aIndex = itemsLength + a
	}
	var bIndex int = b
	if b < 0 {
		bIndex = itemsLength + b
	}
	//check a and b are not negative, and if they are, work with it
	if aIndex >= bIndex {
		if aIndex != a || bIndex != b {
			panic(fmt.Errorf("error end before start[%d:%d]. With original values pass %d:%d", aIndex, bIndex, a, b))
		}
		panic(fmt.Errorf("error end before start[%d:%d]", aIndex, bIndex))
	}
	tss.items = append(tss.items[:aIndex], tss.items[bIndex+1:]...)
}

// get item at index. If a negative number is passed, it will act similar to python, and will get item from index away from the end. Can still throw index out of bounds.
func (tss *ThreadSafeSlice) Get(index int) interface{} {
	tss.lock.RLock()
	defer tss.lock.RUnlock()
	if index >= 0 {
		return tss.items[index]
	}
	return tss.items[len(tss.items)+index]
}
func (tss *ThreadSafeSlice) Set(index int, value interface{}) {
	tss.lock.Lock()
	defer tss.lock.Unlock()
	if index >= 0 {
		tss.items[index] = value
	} else {
		tss.items[len(tss.items)+index] = value
	}
}

func (tss *ThreadSafeSlice) Pop(index int) interface{} {
	tss.lock.Lock()
	defer tss.lock.Unlock()
	if index < 0 {
		//index is negative
		index = len(tss.items) + index
	}
	val := tss.items[index]
	tss.items = append(tss.items[:index], tss.items[index+1:]...)
	return val
}

// gets the length of items in the array at time of testing
func (tss *ThreadSafeSlice) Len() int {
	tss.lock.RLock()
	defer tss.lock.RUnlock()
	return len(tss.items)
}

// Appends an item to the concurrent slice
func (tss *ThreadSafeSlice) Append(item interface{}) {
	tss.lock.Lock()
	defer tss.lock.Unlock()

	tss.items = append(tss.items, item)
}

// do x for each item based on the index and value. Return false to break the loop.
func (tss *ThreadSafeSlice) ForEach(doForEach func(int, interface{}) bool) {
	tss.lock.RLock()
	defer tss.lock.RUnlock()
	for index, value := range tss.items {
		if !doForEach(index, value) {
			return
		}
	}
}
