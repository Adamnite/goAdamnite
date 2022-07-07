package findnode

import (
	"time"

	"github.com/adamnite/go-adamnite/gossip/admnode"
)

// node represents a host on the adamnite network.
type node struct {
	admnode.GossipNode
	addedAt        time.Time
	livenessChecks uint
}

func wrapFindNodes(nodes []*admnode.GossipNode) []*node {
	ret := make([]*node, len(nodes))

	for i, n := range nodes {
		ret[i] = &node{GossipNode: *n}
	}
	return ret
}
