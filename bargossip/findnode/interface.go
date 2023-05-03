package findnode

import "github.com/adamnite/go-adamnite/bargossip/admnode"

// findNodeTransport is implemented by the UDP transports.
type findNodeTransport interface {
	SelfNode() *admnode.GossipNode

	// findSelfNode looks up our own NodeID.
	findSelfNode() []*admnode.GossipNode

	// findRandomNodes looks up a random nodes.
	findRandomNodes() []*admnode.GossipNode

	ping(*admnode.GossipNode) (seq uint64, err error)
}

type FindNode interface {
	Start() bool
	ReplyNodes() []*node
}