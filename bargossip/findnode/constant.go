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
	pingValidateInterval    = 30 * time.Second
	udpLayerResponseTimeout = 8000 * time.Millisecond
	dbUpdateInterval        = 1 * time.Minute
	seedMinTableTime        = 5 * time.Minute

	findNodeRspNodeLimit = 10

	// IP address limits.
	bucketIPLimit, bucketSubnet = 2, 24 // at most 2 addresses from the same /24
	tableIPLimit, tableSubnet   = 10, 24

	// ADMPacket information
	pongRespTimeout    = 500 * time.Millisecond
	ntpFailureThreshold = 32               // Continuous timeouts after which to check NTP
	ntpWarningCooldown  = 10 * time.Minute // Minimum amount of time to pass before repeating NTP warning
	driftThreshold      = 10 * time.Second // Allowed clock drift before warning user

	pingPacketVersion = 1
	hashSize = 256 / 8
	signSize = 512 / 8 + 8 / 8 // signature + packet type
	headSize = hashSize + signSize
)

var (
	headSpace = make([]byte, headSize)
)