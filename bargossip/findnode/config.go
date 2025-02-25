// Adamnite find node module.
// When a node start first time, it must be connect to bootstrap node
// to connect with Adamnite p2p blockchain network.
//

package findnode

import (
	"crypto/ecdsa"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/bargossip/utils"
	"github.com/adamnite/go-adamnite/common/mclock"
)

// Config holds the settings for the findnode listener.
type Config struct {
	// Required settings
	PrivateKey *ecdsa.PrivateKey

	// Optional settings
	PeerBlackList *utils.IPNetList
	PeerWhiteList *utils.IPNetList
	Bootnodes     []*admnode.GossipNode
	Clock         mclock.Clock
}

func (cfg Config) defaults() Config {
	if cfg.Clock == nil {
		cfg.Clock = mclock.System{}
	}
	return cfg
}
