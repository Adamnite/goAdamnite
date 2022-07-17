package findnode

import (
	"context"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

// nodePoolIterator iterates over all seen nodes in node pool.
//
type nodePoolIterator struct {
	nodes    []*node
	ctx      context.Context
	cancel   func()
	nextFind findFunc
	findnode *find
}

type findFunc func(ctx context.Context) *find

func newNodePoolIterator(ctx context.Context, findFunc findFunc) *nodePoolIterator {
	ctx, cancel := context.WithCancel(ctx)
	return &nodePoolIterator{
		ctx:      ctx,
		cancel:   cancel,
		nextFind: findFunc,
	}
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
			it.findnode = it.nextFind(it.ctx)
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
