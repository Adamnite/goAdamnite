package findnode

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	mrand "math/rand"
	"sync"

	"github.com/adamnite/go-adamnite/gossip/admnode"
	"github.com/adamnite/go-adamnite/log15"
)

// NodeTable is the table that stores the neighbor nodes.
type NodeTable struct {
	bootstrapNodes []*node
	db             *admnode.NodeDB
	log            log15.Logger
	transport      findNodeTransport
	rand           *mrand.Rand

	mu sync.Mutex
}

func newNodeTable(findNodeTransport findNodeTransport, db *admnode.NodeDB, bootnodes []*admnode.GossipNode, log log15.Logger) (*NodeTable, error) {
	table := &NodeTable{
		db:        db,
		log:       log,
		transport: findNodeTransport,
	}

	for _, node := range bootnodes {
		if err := node.IsValidate(); err != nil {
			return nil, fmt.Errorf("invalid bootstrap node %q: %v", node, err)
		}
	}

	table.bootstrapNodes = wrapFindNodes(bootnodes)

	table.initialize()
	return table, nil
}

func (tab *NodeTable) initialize() {
	// initialize rand
	tab.resetRand()

	// Load previous nodes stored on database

}

func (tab *NodeTable) resetRand() {
	var b [8]byte
	crand.Read(b[:])

	tab.mu.Lock()
	tab.rand.Seed(int64(binary.BigEndian.Uint64(b[:])))
	tab.mu.Unlock()
}
