package adamdpos

import "time"

const (
	AverageBlockDuration = time.Second * 8  // 8 seconds
	MaxBlockDuration     = time.Second * 16 // 16 seconds

	EpochBlockCount = 162
)

// AdamniteDPOS is a consensus engine based on DPOS implementing.
type AdamniteDPOS struct {
}
