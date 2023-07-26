package safe

import "sync"

//sometimes you need a number to be passed between threads.

type SafeInt struct {
	lock     sync.Mutex
	value    int
	onUpdate chan (int)
}

// a new integer that is safe to use within multiple go routines
func NewSafeInt(value int) *SafeInt {
	return &SafeInt{
		value: value,
	}
}

// returns the sum, as well as storing it locally.
func (tsi *SafeInt) Add(x int) int {
	tsi.lock.Lock()
	defer tsi.lock.Unlock()
	tsi.value += x
	val := tsi.value
	if tsi.onUpdate != nil {
		go func() { tsi.onUpdate <- val }()
	}
	return tsi.value
}

// get the value
func (tsi *SafeInt) Get() int {
	tsi.lock.Lock()
	defer tsi.lock.Unlock()
	return tsi.value
}

// set the value
func (tsi *SafeInt) Set(value int) {
	tsi.lock.Lock()
	defer tsi.lock.Unlock()
	tsi.value = value
	if tsi.onUpdate != nil {
		go func() { tsi.onUpdate <- value }()
	}
}
func (tsi *SafeInt) GetOnUpdate() chan (int) {
	tsi.lock.Lock()
	defer tsi.lock.Unlock()
	if tsi.onUpdate == nil {
		tsi.onUpdate = make(chan int)
	}
	return tsi.onUpdate
}
