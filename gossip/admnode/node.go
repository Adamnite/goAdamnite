package admnode

import (
	"crypto/ecdsa"
	"errors"
)

// GossipNode represents a host on the Adamnite network.
type GossipNode struct {
	id   NodeID
	info NodeInfo
}

// New wraps a node information.
func New(nodeInfo *NodeInfo) (*GossipNode, error) {
	if err := nodeInfo.Verify(); err != nil {
		return nil, err
	}

	node := &GossipNode{info: *nodeInfo}

	node.id = *nodeInfo.GetNodeID()
	if len(node.id) != len(NodeID{}) {
		return nil, errors.New("invalid node id")
	}

	return node, nil
}

func (n *GossipNode) ID() NodeID {
	return n.id
}

func (n *GossipNode) Version() NodeInfoVersion {
	return n.info.GetVersion()
}

func (n *GossipNode) TCP() uint16 {
	return n.info.GetTCP()
}

func (n *GossipNode) UDP() uint16 {
	return n.info.GetUDP()
}

func (n *GossipNode) Pubkey() *ecdsa.PublicKey {
	return n.info.GetPubKey()
}