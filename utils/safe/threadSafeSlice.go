package safe

import (
	"fmt"
	"sort"
	"sync"
)

// a slice type that is thread safe
type SafeSlice struct {
	lock  sync.RWMutex
	items []interface{}
}

func NewSafeSlice() *SafeSlice {
	return &SafeSlice{}
}

func (tss *SafeSlice) Copy() *SafeSlice {
	tss.lock.RLock()
	defer tss.lock.RUnlock()
	newTSS := SafeSlice{
		items: make([]interface{}, len(tss.items)),
		lock:  sync.RWMutex{},
	}
	copy(newTSS.items, tss.items)
	return &newTSS
}
func (tss *SafeSlice) GetItems() []interface{} {
	tss.lock.RLock()
	defer tss.lock.RUnlock()
	return tss.items
}

// pass a lessThan function that takes the indexes in the array
func (tss *SafeSlice) Sort(lessThan func(int, interface{}, int, interface{}) bool) {
	tss.lock.Lock()
	defer tss.lock.Unlock()
	sort.Slice(tss.items, func(i, j int) bool {
		return lessThan(i, tss.items[i], j, tss.items[j])
	})
}

// removes an item from the array based on it's index
func (tss *SafeSlice) Remove(index int) {
	tss.lock.Lock()
	defer tss.lock.Unlock()
	if index < 0 {
		index = len(tss.items) + index
	}
	tss.items = append(tss.items[:index], tss.items[index+1:]...)
}

// removes all the items from a-b including a and b
func (tss *SafeSlice) RemoveFrom(a, b int) {
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
	if aIndex > bIndex {
		if aIndex != a || bIndex != b {
			panic(fmt.Errorf("error end before start[%d:%d]. With original values pass %d:%d", aIndex, bIndex, a, b))
		}
		panic(fmt.Errorf("error end before start[%d:%d]", aIndex, bIndex))
	}
	tss.items = append(tss.items[:aIndex], tss.items[bIndex+1:]...)
}

// get item at index. If a negative number is passed, it will act similar to python, and will get item from index away from the end. Can still throw index out of bounds.
func (tss *SafeSlice) Get(index int) interface{} {
	tss.lock.RLock()
	defer tss.lock.RUnlock()
	if index >= 0 {
		return tss.items[index]
	}
	return tss.items[len(tss.items)+index]
}
func (tss *SafeSlice) Set(index int, value interface{}) {
	tss.lock.Lock()
	defer tss.lock.Unlock()
	if index >= 0 {
		tss.items[index] = value
	} else {
		tss.items[len(tss.items)+index] = value
	}
}

func (tss *SafeSlice) Pop(index int) interface{} {
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
func (tss *SafeSlice) Len() int {
	tss.lock.RLock()
	defer tss.lock.RUnlock()
	return len(tss.items)
}

// Appends an item to the concurrent slice
func (tss *SafeSlice) Append(item interface{}) {
	tss.lock.Lock()
	defer tss.lock.Unlock()

	tss.items = append(tss.items, item)
}

// do x for each item based on the index and value. Return false to break the loop.
func (tss *SafeSlice) ForEach(doForEach func(int, interface{}) bool) {
	tss.lock.RLock()
	defer tss.lock.RUnlock()
	for index, value := range tss.items {
		if !doForEach(index, value) {
			return
		}
	}
}
