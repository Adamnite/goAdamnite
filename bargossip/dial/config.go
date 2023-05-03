package dial

import (
	crand "crypto/rand"
	"encoding/binary"
	mrand "math/rand"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/bargossip/utils"
	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/log15"
)

type Config struct {
	// SelfID is the id of local node
	SelfID admnode.NodeID

	// PeerBlackList is the black list of the peers
	PeerBlackList *utils.IPNetList

	// PeerWhiteList is the white list of the peers
	PeerWhiteList *utils.IPNetList

	// MaxOutboundConnection is the maximum number of outbounding peer connections
	MaxOutboundConnections int

	// MaxPendingOutboundConnections is the maximum number of pending outbound peer connections
	MaxPendingOutboundConnections int

	Log log15.Logger

	Clock mclock.Clock

	Rand *mrand.Rand
}

func (cfg Config) WithDefaults() Config {
	if cfg.MaxOutboundConnections == 0 {
		cfg.MaxOutboundConnections = defaultMaxOutboundConnections
	}
	if cfg.MaxPendingOutboundConnections == 0 {
		cfg.MaxPendingOutboundConnections = defaultMaxPendingOutboundConnections
	}
	if cfg.Log == nil {
		cfg.Log = log15.Root()
	}
	if cfg.Clock == nil {
		cfg.Clock = mclock.System{}
	}

	if cfg.Rand == nil {
		seedb := make([]byte, 8)
		crand.Read(seedb)
		seed := int64(binary.BigEndian.Uint64(seedb))
		cfg.Rand = mrand.New(mrand.NewSource(seed))
	}
	return cfg
}
