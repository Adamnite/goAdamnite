package types

import "github.com/adamnite/go-adamnite/utils"

type writeCounter utils.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}
