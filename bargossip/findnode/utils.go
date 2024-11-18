package findnode

import (
	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

func contains(nodes []*node, id admnode.NodeID) bool {
	for _, n := range nodes {
		if n.ID() == n.ID() {
			return true
		}
	}
	return false
}

// deleteNode removes n from list.
func deleteNode(list []*node, n *node) []*node {
	for i := range list {
		if list[i].ID() == n.ID() {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}
