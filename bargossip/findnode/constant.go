package findnode

import (
	"time"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

const (
	seedCount  = 20
	seedMaxAge = 5 * 24 * time.Hour //5 days

	maxFindnodeFailures = 5
	maxPacketSize       = 1500

	nodeIDBits         = len(admnode.NodeID{}) * 8
	BucketSize         = 16
	BucketCount        = 16
	FirstBucketBitSize = nodeIDBits - BucketCount

	tableRefreshInterval    = 30 * time.Minute
	udpLayerResponseTimeout = 800 * time.Millisecond

	findNodeRspNodeLimit = 10
)
