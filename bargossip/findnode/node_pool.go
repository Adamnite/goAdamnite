package findnode

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	mrand "math/rand"
	"sync"
	"time"
	"net"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/bargossip/utils"
	"github.com/adamnite/go-adamnite/log15"
)

// NodePool is the table that stores the neighbor nodes.
type NodePool struct {
	bootstrapNodes []*node
	db             *admnode.NodeDB
	log            log15.Logger
	transport      findNodeTransport
	rand           *mrand.Rand
	buckets        [BucketCount]*bucket

	ips     	   utils.DistinctNetSet

	mu sync.Mutex

	closed   chan struct{}
	closeReq chan struct{}
}

type bucket struct {
	whitelist []*node
	graylist  []*node
	ips       utils.DistinctNetSet
}

func newNodePool(findNodeTransport findNodeTransport, db *admnode.NodeDB, bootnodes []*admnode.GossipNode, log log15.Logger) (*NodePool, error) {
	table := &NodePool{
		db:        db,
		log:       log,
		transport: findNodeTransport,
		rand:      mrand.New(mrand.NewSource(0)),
		ips:       utils.DistinctNetSet{Subnet: tableSubnet, Limit: tableIPLimit},
	}

	for i := range table.buckets {
		table.buckets[i] = &bucket{
			ips: utils.DistinctNetSet{Subnet: bucketSubnet, Limit: bucketIPLimit},
		}
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
	var(
		refreshTick 		= time.NewTicker(tableRefreshInterval)
		pingValidateTick	= time.NewTicker(tab.nextPingValidateTime())
		dbValidateTick      = time.NewTicker(dbUpdateInterval)

		refreshDone 		= make(chan struct{})
		pingValidateDone chan struct{}
	)

	defer refreshTick.Stop()
	defer pingValidateTick.Stop()

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
		case <-pingValidateTick.C:
			pingValidateDone = make(chan struct{})
			go tab.doPingValidate(pingValidateDone)
		case <-pingValidateDone:
			pingValidateTick.Reset(tab.nextPingValidateTime())
			pingValidateDone = nil
		case <-dbValidateTick.C:
			go tab.updateLiveNodesOnDB()
		case <-tab.closeReq:
			break loop
		}
	}

	if refreshDone != nil {
		<-refreshDone
	}

	close(tab.closed)
}

// updateLiveNodesOnDB adds nodes from the table to the database if they have been in the table
// longer then minTableTime.
func (tab *NodePool) updateLiveNodesOnDB() {
	tab.mu.Lock()
	defer tab.mu.Unlock()

	now := time.Now()
	for _, b := range &tab.buckets {
		for _, n := range b.whitelist {
			if n.livenessChecks > 0 && now.Sub(n.addedAt) >= seedMinTableTime {
				tab.db.UpdateNode(unWrapFindNode(n))
			}
		}
	}
}

func (tab *NodePool) doPingValidate(done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	pingNode, bucketIndex := tab.getNodesWithPingValidate()
	if pingNode == nil {
		return
	}

	// Ping the selected node and wait for a pong message.
	_, err := tab.transport.ping(unWrapFindNode(pingNode))

	// if pingNode.Seq() < remoteSeq {
		
	// }

	tab.mu.Lock();
	defer tab.mu.Unlock();

	bucket := tab.buckets[bucketIndex]
	
	if err != nil {
		if r := tab.replaceNode(bucket, pingNode); r != nil {
			tab.log.Debug("Replaced dead node", "bucket", bucketIndex, "id", pingNode.ID(), "ip", pingNode.IP(), "checks", pingNode.livenessChecks, "replaceID", r.ID(), "replaceIP", r.IP())
		} else {
			tab.log.Debug("Removed dead node", "bucket", bucketIndex, "id", pingNode.ID(), "ip", pingNode.IP(), "checks", pingNode.livenessChecks)
		}
	}

	pingNode.livenessChecks++
	tab.log.Debug("Revalidated node", "bucket", bucketIndex, "id", pingNode.ID(), "ip", pingNode.IP(), "checks", pingNode.livenessChecks)
	tab.bumpupInBucket(bucket, pingNode)
}

// bumpupInBucket moves the given node to the front of the bucket entry list
// if it is contained in that list.
func (tab *NodePool) bumpupInBucket(b *bucket, n *node) bool {
	for i := range b.whitelist {
		if b.whitelist[i].ID() == n.ID() {
			if !n.IP().Equal(b.whitelist[i].IP()) {
				// Endpoint has changed, ensure that the new IP fits into table limits.
				tab.removeIP(b, b.whitelist[i].IP())
				if !tab.addIP(b, n.IP()) {
					// It doesn't, put the previous one back.
					tab.addIP(b, b.whitelist[i].IP())
					return false
				}
			}
			// Move it to the front.
			copy(b.whitelist[1:], b.whitelist[:i])
			b.whitelist[0] = n
			return true
		}
	}
	return false
}

func (tab *NodePool) addIP(b *bucket, ip net.IP) bool {
	if utils.IsLAN(ip) {
		return true
	}
	if !tab.ips.Add(ip) {
		tab.log.Debug("IP exceeds table limit", "ip", ip)
		return false
	}
	if !b.ips.Add(ip) {
		tab.log.Debug("IP exceeds bucket limit", "ip", ip)
		tab.ips.Remove(ip)
		return false
	}
	return true
}

// replaceNode removes node from the graylist.
func (tab *NodePool) replaceNode(b *bucket, n *node) *node {
	if len(b.whitelist) == 0 || b.whitelist[len(b.whitelist) - 1].ID() != n.ID() {
		return nil
	}

	if len(b.graylist) == 0 {
		tab.deleteInBucket(b, n)
		return nil
	}

	r := b.graylist[tab.rand.Intn(len(b.graylist))]
	b.graylist = deleteNode(b.graylist, r)
	b.whitelist[len(b.whitelist) - 1] = r
	tab.removeIP(b, n.IP())
	return r
}

func (tab *NodePool) removeIP(b *bucket, ip net.IP) {
	if utils.IsLAN(ip) {
		return
	}
	tab.ips.Remove(ip)
	b.ips.Remove(ip)
}

func (tab *NodePool) getNodesWithPingValidate() (*node, int) {
	tab.mu.Lock()
	defer tab.mu.Unlock()

	for _, index := range tab.rand.Perm(len(tab.buckets)) {
		bucket := tab.buckets[index]

		if len(bucket.whitelist) > 0 {
			last := bucket.whitelist[len(bucket.whitelist) - 1]
			return last, index
		}
	}

	return nil, 0
}

func (tab *NodePool) nextPingValidateTime() time.Duration {
	tab.mu.Lock()
	defer tab.mu.Unlock()

	return time.Duration(tab.rand.Int63n(int64(pingValidateInterval)))
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

func (tab *NodePool) addNode(node *node) {
	tab.mu.Lock()
	defer tab.mu.Unlock()
	tab.addInBucket(tab.getBucket(*node.ID()), node)
}

func (tab *NodePool) addInBucket(b *bucket, n *node) {
	b.whitelist = addNode(b.whitelist, n)
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