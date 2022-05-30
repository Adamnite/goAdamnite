package core

import "time"

type TxPoolConfig struct {
	Lifetime time.Duration // Maximum amount of time non-executable transaction are queued.

	AccountSlots uint64
	AccountQueue uint64
	GlobalSlots  uint64
	GlobalQueue  uint64
}

var DefaultTxPoolConfig = TxPoolConfig{
	Lifetime: 1 * time.Hour,

	AccountSlots: 32,
	AccountQueue: 64,
	GlobalSlots:  8192,
	GlobalQueue:  1024,
}
