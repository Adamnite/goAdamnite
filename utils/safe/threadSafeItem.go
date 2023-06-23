package safe

import "sync"

type SafeItem struct {
	lock sync.RWMutex
	item any
}

func NewSafeItem(item any) *SafeItem {
	return &SafeItem{
		lock: sync.RWMutex{},
		item: item,
	}
}

func (si *SafeItem) Get() any {
	si.lock.RLock()
	defer si.lock.RUnlock()
	return si.item
}
func (si *SafeItem) Set(item any) {
	si.lock.Lock()
	defer si.lock.Unlock()
	si.item = item
}

func GetItem[T any](si *SafeItem) T {
	return si.Get().(T)
}
