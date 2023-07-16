package findnode

import (
	"context"

	"github.com/adamnite/go-adamnite/bargossip/admnode"

	log "github.com/sirupsen/logrus"
)

type nodeFindFunc func(*node) ([]*node, error)

// find performs a network search for nodes close to the given target node.
type find struct {
	table    *NodePool
	findFunc nodeFindFunc
	result   nodes
	queries  int

	asked, seen map[admnode.NodeID]bool
	replyNodes  []*node

	// channels
	cancelCh <-chan struct{}
	replyCh  chan []*node
}

func newFind(ctx context.Context, table *NodePool, targetNodeID admnode.NodeID, findFunc nodeFindFunc) *find {
	f := &find{
		table:    table,
		findFunc: findFunc,
		result:   nodes{targetId: targetNodeID},
		queries:  -1,
		asked:    make(map[admnode.NodeID]bool),
		seen:     make(map[admnode.NodeID]bool),
		cancelCh: ctx.Done(),
		replyCh:  make(chan []*node, 3),
	}

	f.asked[*table.transport.SelfNode().ID()] = true
	return f
}

func (f *find) run() []*admnode.GossipNode {
	for f.start() {
	}
	return unWrapFindNodes(f.result.nodes)
}

func (f *find) start() bool {
	for f.startQueries() {
		select {
		case nodes := <-f.replyCh:
			f.replyNodes = f.replyNodes[:0]
			for _, n := range nodes {
				if n != nil && !f.seen[*n.ID()] {
					f.seen[*n.ID()] = true
					f.result.push(n, BucketSize)
					f.replyNodes = append(f.replyNodes, n)
				}
			}
			f.queries--
			if len(f.replyNodes) > 0 {
				return true
			}
		case <-f.cancelCh:
			f.close()
		}
	}
	return false
}

func (f *find) startQueries() bool {
	if f.findFunc == nil {
		return false
	}

	if f.queries == -1 {
		closet := f.table.findnodeByID(f.result.targetId, BucketSize, false)
		f.queries = 1
		f.replyCh <- closet.nodes
		return true
	}

	for i := 0; i < len(f.result.nodes) && f.queries < 3; i++ {
		node := f.result.nodes[i]
		if !f.asked[*node.ID()] {
			f.asked[*node.ID()] = true
			f.queries++
			go f.query(node)
		}
	}

	return f.queries > 0
}

func (f *find) query(n *node) {
	fails := f.table.db.FindFails(*n.ID(), n.IP())
	nodes, err := f.findFunc(n)
	if err == errClosed {
		f.replyCh <- nil
		return
	} else if len(nodes) == 0 {
		fails++
		f.table.db.UpdateFindFails(*n.ID(), n.IP(), fails)

		dropped := false
		if fails > maxFindnodeFailures && f.table.getBucketLen(*n.ID()) >= BucketSize/2 {
			f.table.deleteNode(n)
			dropped = true
		}
		log.Trace("Findnode failed", "id", n.ID(), "failedcount", fails, "dropped", dropped, "err", err)
	}

	for _, node := range nodes {
		f.table.addSeenNode(node)
	}

	if fails > 0 {
		f.table.db.UpdateFindFails(*n.ID(), n.IP(), 0)
	}
	f.replyCh <- nodes
}

func (f *find) close() {
	for f.queries > 0 {
		<-f.replyCh
		f.queries--
	}
	f.findFunc = nil
	f.replyNodes = nil
}
