package findnode

import (
	"net"
	"sort"
	"time"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
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
		ret[i] = wrapFindNode(n)
	}
	return ret
}

func unWrapFindNodes(nodes []*node) []*admnode.GossipNode {
	ret := make([]*admnode.GossipNode, len(nodes))
	for i, n := range nodes {
		ret[i] = unWrapFindNode(n)
	}
	return ret
}

func wrapFindNode(n *admnode.GossipNode) *node {
	return &node{GossipNode: *n}
}

func unWrapFindNode(node *node) *admnode.GossipNode {
	return &node.GossipNode
}

func (n *node) getUDPAddr() net.UDPAddr {
	return net.UDPAddr{IP: n.IP(), Port: int(n.UDP())}
}

// nodes is a list of nodes, ordered by distance to target.
type nodes struct {
	nodes    []*node
	targetId admnode.NodeID
}

func (ns *nodes) push(n *node, maxNodeCount int) {
	index := sort.Search(len(ns.nodes), func(i int) bool {
		return admnode.DistanceCmp(*ns.nodes[i].ID(), *n.ID(), ns.targetId) > 0
	})
	if index > 0 {
		if index == len(ns.nodes) {
			ns.nodes = append(ns.nodes, n)
		} else {
			copy(ns.nodes[index+1:], ns.nodes[index:])
			ns.nodes[index] = n
		}

	} else {
		ns.nodes = append(ns.nodes, n)
	}
}
