package findnode

import (
	"context"
	crand "crypto/rand"
	"net"
	"sync"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/log15"
)

// FindNode is the implementation of the node discovery protocol.
type FindNode struct {
	conn      *net.UDPConn
	localNode *admnode.LocalNode
	nodeTable *NodeTable
	log       log15.Logger
	clock     mclock.Clock

	// terminate items
	closeCtx       context.Context
	cancelCloseCtx context.CancelFunc
	wg             sync.WaitGroup
	closeOnce      sync.Once
}

// Start runs the findnode module and listen the connection.
func Start(conn *net.UDPConn, localNode *admnode.LocalNode, cfg Config) (*FindNode, error) {
	closeCtx, cancelCloseCtx := context.WithCancel(context.Background())
	cfg = cfg.defaults()

	findNode := &FindNode{
		conn:      conn,
		localNode: localNode,
		log:       cfg.Log,
		clock:     cfg.Clock,
		// terminate items
		closeCtx:       closeCtx,
		cancelCloseCtx: cancelCloseCtx,
	}

	table, err := newNodeTable(findNode, findNode.localNode.Database(), cfg.Bootnodes, cfg.Log)
	if err != nil {
		return nil, err
	}

	findNode.nodeTable = table

	go findNode.nodeTable.backgroundThread()

	return findNode, nil
}

func (n *FindNode) SelfNode() *admnode.GossipNode {
	return n.localNode.Node()
}

func (n *FindNode) findSelfNode() []*admnode.GossipNode {
	return newFind(n.closeCtx, n.nodeTable, *n.SelfNode().ID(), func(nd *node) ([]*node, error) {
		return n.findFunc(nd, *n.SelfNode().ID())
	}).run()
}

func (n *FindNode) findRandomNodes() []*admnode.GossipNode {
	var target admnode.NodeID
	crand.Read(target[:])
	return newFind(n.closeCtx, n.nodeTable, target, func(nd *node) ([]*node, error) {
		return n.findFunc(nd, target)
	}).run()
}

func (n *FindNode) findFunc(destNode *node, target admnode.NodeID) ([]*node, error) {
	var (
		dists         = n.findDistances(target, *destNode.ID())
		distanceNodes = nodes{targetId: target}
	)

	resNodes, err := n.findNode(&destNode.GossipNode, dists)
	if err == errClosed {
		return nil, err
	}
	for _, wnode := range resNodes {
		if wnode.ID() != n.SelfNode().ID() {
			distanceNodes.push(&node{GossipNode: *wnode}, 16)
		}
	}
	return distanceNodes.nodes, err
}

func (n *FindNode) findDistances(target, dest admnode.NodeID) (dists []uint) {
	distance := admnode.LogDist(target, dest)
	dists = append(dists, uint(distance))

	for i := 1; len(dists) < 3; i++ {
		if distance+i < 256 {
			dists = append(dists, uint(distance+i))
		}
		if distance-i > 0 {
			dists = append(dists, uint(distance-i))
		}
	}
	return dists
}

func (n *FindNode) findNode(node *admnode.GossipNode, distances []uint) ([]*admnode.GossipNode, error) {
	resp := n.call(node, admpacket.RspFindnodeMsg, admpacket.Findnode{})
}

func (n *FindNode) call(node *admnode.GossipNode, responseType byte, packet admpacket.Packet) {}
