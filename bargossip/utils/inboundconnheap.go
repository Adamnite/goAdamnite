package utils

import (
	"container/heap"

	"github.com/adamnite/go-adamnite/utils/mclock"
)

type inboundConnItem struct {
	item       string
	expireTime mclock.AbsTime
}

type InboundConnHeap []inboundConnItem

func (h *InboundConnHeap) Push(x interface{}) { *h = append(*h, x.(inboundConnItem)) }
func (h *InboundConnHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
func (h InboundConnHeap) Len() int           { return len(h) }
func (h InboundConnHeap) Less(i, j int) bool { return h[i].expireTime < h[j].expireTime }
func (h InboundConnHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *InboundConnHeap) Add(item string, expireTime mclock.AbsTime) {
	heap.Push(h, inboundConnItem{item, expireTime})
}

func (h InboundConnHeap) Contains(item string) bool {
	for _, v := range h {
		if v.item == item {
			return true
		}
	}
	return false
}

func (h *InboundConnHeap) nextExpiry() mclock.AbsTime {
	return (*h)[0].expireTime
}

func (h *InboundConnHeap) Expire(now mclock.AbsTime, onExp func(string)) {
	for h.Len() > 0 && h.nextExpiry() < now {
		item := heap.Pop(h) // remove oldest one
		if onExp != nil {
			onExp(item.(inboundConnItem).item)
		}
	}
}
