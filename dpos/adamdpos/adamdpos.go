package adamdpos

import "time"
import "fmt"
import "units"

const (
	AverageBlockDuration = time.Second * 1  // 1 second
	MaxBlockDuration     = time.Second * 10 // 10 seconds
	BlockSize 	     = units.MegaByte * 5 // 5 megabytes is the block size limit
	

	EpochBlockCount = 162
)

// AdamniteDPOS is a consensus engine based on DPOS implementing.
type AdamniteDPOS struct {
}
