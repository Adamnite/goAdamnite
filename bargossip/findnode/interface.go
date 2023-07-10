package findnode

import "github.com/adamnite/go-adamnite/bargossip/admnode"

type findNodeTransport interface {
	SelfNode() *admnode.GossipNode

	// findSelfNode looks up our own NodeID.
	findSelfNode() []*admnode.GossipNode

	// findRandomNodes looks up a random nodes.
	findRandomNodes() []*admnode.GossipNode
}
