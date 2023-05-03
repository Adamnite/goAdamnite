package findnode

import (
	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

func contains(nodes []*node, id admnode.NodeID) bool {
	for _, n := range nodes {
		if *n.ID() == id {
			return true
		}
	}
	return false
}

func addNode(list []*node, n *node) []*node {
	// check if node exists on the list
	for i := range list {
		if *list[i].ID() == *n.ID() {
			return list
		}
	}
	return append(list, n)
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
