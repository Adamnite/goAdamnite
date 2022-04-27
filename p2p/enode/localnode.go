package enode

import (
	"crypto/ecdsa"
	"net"
	"sync"
	"sync/atomic"

	"github.com/adamnite/go-adamnite/p2p/netutil"
)

type localNodeEndpoint struct {
	track       *netutil.IPTracker
	staticIP    net.IP
	fallbackIP  net.IP
	fallbackUDP int
}

type LocalNode struct {
	id          ID
	key         *ecdsa.PrivateKey
	db          *DB
	currentNode atomic.Value

	mu        sync.Mutex
	seq       uint64
	endpoint4 localNodeEndpoint
	endpoint6 localNodeEndpoint
}

func NewLocalNode(db *DB, key *ecdsa.PrivateKey) *LocalNode {
	localNode := &LocalNode{}
	return localNode
}
