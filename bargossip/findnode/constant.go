package findnode

import (
	"time"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

const (
	seedCount  = 20
	seedMaxAge = 5 * 24 * time.Hour

	maxFindnodeFailures = 5

	nodeIDBits         = len(admnode.NodeID{}) * 8
	BucketSize         = 16
	BucketCount        = 16
	FirstBucketBitSize = nodeIDBits - BucketCount

	tableRefreshInterval = 30 * time.Minute
)
