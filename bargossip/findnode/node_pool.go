package findnode

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	mrand "math/rand"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

// NodePool is the table that stores the neighbor nodes.
type NodePool struct {
	bootstrapNodes []*node
	db             *admnode.NodeDB
	log            log15.Logger
	transport      findNodeTransport
	rand           *mrand.Rand
	buckets        [BucketCount]*bucket

	mu sync.Mutex

	closed   chan struct{}
	closeReq chan struct{}
}

type bucket struct {
	whitelist []*node
	graylist  []*node
}

func newNodePool(findNodeTransport findNodeTransport, db *admnode.NodeDB, bootnodes []*admnode.GossipNode, log log15.Logger) (*NodePool, error) {
	table := &NodePool{
		db:        db,
		log:       log,
		transport: findNodeTransport,
		rand:      mrand.New(mrand.NewSource(0)),
	}

	for i := range table.buckets {
		table.buckets[i] = &bucket{}
	}

	for _, node := range bootnodes {
		if err := node.IsValidate(); err != nil {
			return nil, fmt.Errorf("invalid bootstrap node %v: %v", node, err)
		}
	}

	table.bootstrapNodes = wrapFindNodes(bootnodes)

	table.initialize()
	return table, nil
}

func (tab *NodePool) initialize() {
	// initialize rand
	tab.resetRand()

	// Load previous nodes stored on database
	tab.loadNodes()
}

func (tab *NodePool) resetRand() {
	var b [8]byte
	crand.Read(b[:])

	tab.mu.Lock()
	tab.rand.Seed(int64(binary.BigEndian.Uint64(b[:])))
	tab.mu.Unlock()
}

func (tab *NodePool) loadNodes() {
	seeds := wrapFindNodes(tab.db.QueryRandomNodes(seedCount, seedMaxAge))
	seeds = append(seeds, tab.bootstrapNodes...)
	for i := range seeds {
		seed := seeds[i]
		tab.addSeenNode(seed)
		tab.log.Trace("Found seed node", "id", seed.ID(), "addr", seed.getUDPAddr())
	}
}

func (tab *NodePool) addSeenNode(n *node) {
	if n.ID() == tab.transport.SelfNode().ID() {
		return
	}

	tab.mu.Lock()
	defer tab.mu.Unlock()

	b := tab.getBucket(*n.ID())
	if contains(b.whitelist, *n.ID()) {
		return
	}

	if len(b.whitelist) >= BucketSize {
		return
	}

	b.whitelist = append(b.whitelist, n)
	b.graylist = deleteNode(b.graylist, n)
	n.addedAt = time.Now()
}

func (tab *NodePool) getBucket(id admnode.NodeID) *bucket {
	distance := admnode.LogDist(*tab.transport.SelfNode().ID(), id)
	return tab.getBucketAtDistance(distance)
}

func (tab *NodePool) getBucketAtDistance(distance int) *bucket {
	if distance <= FirstBucketBitSize {
		return tab.buckets[0]
	}
	return tab.buckets[distance-FirstBucketBitSize-1]
}

// backgroundThread update buckets
func (tab *NodePool) backgroundThread() {
	var refreshTick = time.NewTicker(tableRefreshInterval)
	var refreshDone = make(chan struct{})

	defer refreshTick.Stop()

	go tab.refreshTable(refreshDone)

loop:
	for {
		select {
		case <-refreshTick.C:
			tab.resetRand()
			if refreshDone == nil {
				refreshDone = make(chan struct{})
				go tab.refreshTable(refreshDone)
			}
		case <-refreshDone:
			refreshDone = nil
		case <-tab.closeReq:
			break loop
		}
	}

	if refreshDone != nil {
		<-refreshDone
	}

	close(tab.closed)
}

// refreshTable performs a finding a random node to keep bucket full.
func (tab *NodePool) refreshTable(done chan struct{}) {
	defer close(done)

	tab.loadNodes()
	tab.transport.findSelfNode()

	for i := 0; i < 3; i++ {
		tab.transport.findRandomNodes()
	}
}

func (tab *NodePool) close() {
	close(tab.closeReq)
	<-tab.closed
}

func (tab *NodePool) findnodeByID(target admnode.NodeID, count int, live bool) *nodes {
	tab.mu.Lock()
	defer tab.mu.Unlock()

	ns := &nodes{targetId: target}
	liveNodes := &nodes{targetId: target}

	for _, bucket := range tab.buckets {
		for _, node := range bucket.whitelist {
			ns.push(node, count)
			if live && node.livenessChecks > 0 {
				liveNodes.push(node, count)
			}
		}
	}

	if live && len(liveNodes.nodes) > 0 {
		return liveNodes
	}
	return ns
}

func (tab *NodePool) getBucketLen(id admnode.NodeID) int {
	tab.mu.Lock()
	defer tab.mu.Unlock()
	return len(tab.getBucket(id).whitelist)
}

func (tab *NodePool) deleteNode(node *node) {
	tab.mu.Lock()
	defer tab.mu.Unlock()
	tab.deleteInBucket(tab.getBucket(*node.ID()), node)
}

func (tab *NodePool) deleteInBucket(b *bucket, n *node) {
	b.whitelist = deleteNode(b.whitelist, n)
}

// getNode returns the node with the given ID on whitelist of table.
func (tab *NodePool) getNode(id admnode.NodeID) *admnode.GossipNode {
	tab.mu.Lock()
	defer tab.mu.Unlock()

	bucket := tab.getBucket(id)
	for _, n := range bucket.whitelist {
		if *n.ID() == id {
			return &n.GossipNode
		}
	}

	return nil
}
