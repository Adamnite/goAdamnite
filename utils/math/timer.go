package utils

import (
	"encoding/binary"
	"time"
)

// Timer represents a generic timer interface.
type Timer interface {
	Timeout() bool
	Encode() []byte
	Decode(data []byte) error
	Zero()
}

// MonotonicTimer implements a timer that measures elapsed time using monotonic clock.
type MonotonicTimer struct {
	startTime time.Time
	duration  time.Duration
}

// NewMonotonicTimer creates a new monotonic timer with the specified duration.
func NewMonotonicTimer(duration time.Duration) *MonotonicTimer {
	return &MonotonicTimer{
		startTime: time.Now(),
		duration:  duration,
	}
}

// Timeout checks if the monotonic timer has expired.
func (t *MonotonicTimer) Timeout() bool {
	return time.Since(t.startTime) >= t.duration
}

// Encode encodes the monotonic timer into a byte slice.
func (t *MonotonicTimer) Encode() []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(t.duration))
	return data
}

// Decode decodes the monotonic timer from a byte slice.
func (t *MonotonicTimer) Decode(data []byte) error {
	if len(data) != 8 {
		return ErrInvalidData
	}
	duration := time.Duration(binary.BigEndian.Uint64(data))
	t.startTime = time.Now()
	t.duration = duration
	return nil
}

// Zero resets the monotonic timer.
func (t *MonotonicTimer) Zero() {
	t.startTime = time.Time{}
	t.duration = 0
}

// FrozenTimer implements a timer that remains frozen at a specific time.
type FrozenTimer struct {
	frozenTime time.Time
}

// NewFrozenTimer creates a new frozen timer with the specified time.
func NewFrozenTimer(frozenTime time.Time) *FrozenTimer {
	return &FrozenTimer{
		frozenTime: frozenTime,
	}
}

// Timeout always returns false for a frozen timer.
func (t *FrozenTimer) Timeout() bool {
	return false
}

// Encode encodes the frozen timer into a byte slice.
func (t *FrozenTimer) Encode() []byte {
	return t.frozenTime.UTC().MarshalBinary()
}

// Decode decodes the frozen timer from a byte slice.
func (t *FrozenTimer) Decode(data []byte) error {
	frozenTime := time.Time{}
	if err := frozenTime.UnmarshalBinary(data); err != nil {
		return err
	}
	t.frozenTime = frozenTime.UTC()
	return nil
}

// Zero resets the frozen timer.
func (t *FrozenTimer) Zero() {
	t.frozenTime = time.Time{}
}
