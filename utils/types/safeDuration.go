package utils

import (
	"sync"
	"time"
)

type SafeDuration struct {
    d  time.Duration
    mu sync.RWMutex
}

func NewSafeDuration(dur time.Duration) *SafeDuration {
	return &SafeDuration{
		d: dur,
	}
}

func (d *SafeDuration) Duration() time.Duration {
    d.mu.RLock()
    defer d.mu.RUnlock()
    return d.d
}

func (d *SafeDuration) SetDuration(dur time.Duration) {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.d = dur
}