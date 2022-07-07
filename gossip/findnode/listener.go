package findnode

import (
	"context"
	"net"
	"sync"

	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/gossip/admnode"
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

	return findNode, nil
}
