// A BAR Resilient P2P network server for distributed ledger system

package gossip

import (
	"errors"
	"net"
	"sync"

	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/gossip/admnode"
	"github.com/adamnite/go-adamnite/gossip/nat"
	"github.com/adamnite/go-adamnite/log15"
)

// Server manages the peer connections.
type Server struct {
	Config

	isRunning bool

	listener net.Listener
	log      log15.Logger

	localnode *admnode.LocalNode

	lock   sync.Mutex
	loopWG sync.WaitGroup

	nodedb *admnode.NodeDB

	// Channels
	quit chan struct{}
}

// Start starts the server.
func (srv *Server) Start() (err error) {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	if srv.isRunning {
		return errors.New("adamnite p2p server is running")
	}

	srv.isRunning = true
	if err = srv.initialize(); err != nil {
		return err
	}

	srv.loopWG.Add(1)
	go srv.run()
	return nil
}

func (srv *Server) initialize() (err error) {
	srv.log = srv.Config.Logger
	if srv.log == nil {
		srv.log = log15.Root()
	}

	if srv.clock == nil {
		srv.clock = mclock.System{}
	}

	if srv.ListenAddr == "" {
		srv.log.Warn("adamnite p2p server listening address is not set")
	}

	if srv.ServerPrvKey == nil {
		return errors.New("adaminte p2p server private key must be set to a non-nil key")
	}

	srv.quit = make(chan struct{})

	if err := srv.initializeLocalNode(); err != nil {
		return err
	}

	return nil
}

// initializeLocalNode
func (srv *Server) initializeLocalNode() error {
	// Create the local node DB.
	db, err := admnode.OpenDB(srv.Config.NodeDatabase)
	if err != nil {
		return err
	}

	srv.nodedb = db
	srv.localnode = admnode.NewLocalNode(db, srv.ServerPrvKey)

	switch srv.NAT.(type) {
	case nil:
	case nat.ExtIP:
		ip, _ := srv.NAT.ExternalIP()
		srv.localnode.SetIP(ip)
	default:
		srv.loopWG.Add(1)
		go func() {
			defer srv.loopWG.Done()
			if ip, err := srv.NAT.ExternalIP(); err == nil {
				srv.localnode.SetIP(ip)
			}
		}()
	}
	return nil
}

func (srv *Server) run() {
	srv.log.Info("Adamnite p2p server started", "localnode", srv.localnode.NodeInfo().ToURL())
	defer srv.loopWG.Done()
}
