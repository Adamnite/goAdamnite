package findnode

import (
	"context"
	crand "crypto/rand"
	"time"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

// nodePoolIterator iterates over all seen nodes in node pool.
//
type nodePoolIterator struct {
	nodes    []*node
	ctx      context.Context
	udpLayer *UDPLayer
	cancel   func()
	findnode *find
	init     bool
}

type findFunc func(ctx context.Context) *find

func NewNodePoolIterator(udpL *UDPLayer) *nodePoolIterator {
	ctx, cancel := context.WithCancel(context.Background())

	return &nodePoolIterator{
		ctx:      ctx,
		cancel:   cancel,
		udpLayer: udpL,
		init:     false,
	}
}

func (it *nodePoolIterator) nextFind(ctx context.Context, udpLayer *UDPLayer) *find {
	var target admnode.NodeID
	crand.Read(target[:])

	findNode := newFind(udpLayer.closeCtx, udpLayer.nodeTable, target, func(nd *node) ([]*node, error) {
		return udpLayer.findFunc(nd, target)
	})
	return findNode
}

func (it *nodePoolIterator) Node() *admnode.GossipNode {
	if len(it.nodes) == 0 {
		return nil
	}
	return &it.nodes[0].GossipNode
}

func (it *nodePoolIterator) Next() bool {
	if len(it.nodes) > 0 {
		it.nodes = it.nodes[1:]
	}
	for len(it.nodes) == 0 {
		if it.ctx.Err() != nil {
			it.nodes = nil
			return false
		}
		if it.findnode == nil {
			it.findnode = it.nextFind(it.ctx, it.udpLayer)

			if it.init == true {
				timer := time.NewTimer(30 * time.Second)
				<-timer.C
			} else {
				it.init = true
			}
		
			continue
		}
		if !it.findnode.start() {
			it.findnode = nil
			continue
		}
		it.nodes = it.findnode.replyNodes
	}
	return true
}

func (it *nodePoolIterator) Close() {
	it.cancel()
}
