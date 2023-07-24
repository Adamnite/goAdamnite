package types

import "github.com/adamnite/go-adamnite/utils"
import "github.com/adamnite/go-adamnite/utils/bytes"

type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}
