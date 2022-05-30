package node

import (
	"path/filepath"
	"sync"

	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/p2p"
	"github.com/adamnite/go-adamnite/rpc"
)

const (
	initializingState = iota
	runningState
	closedState
)

type Node struct {
	config *Config
	log    log15.Logger
	stop   chan struct{}

	ipc    *ipcServer
	server *p2p.Server

	startStopLock sync.Mutex
	state         int
	lock          sync.Mutex

	inprocHandler *rpc.Server
	rpcAPIs       []rpc.API
}

func New(cfg *Config) (*Node, error) {
	confCopy := *cfg
	cfg = &confCopy

	if cfg.DataDir != "" {
		absdatadir, err := filepath.Abs(cfg.DataDir)
		if err != nil {
			return nil, err
		}

		cfg.DataDir = absdatadir
	}

	if cfg.Logger == nil {
		cfg.Logger = log15.New()
	}

	node := &Node{
		config:        cfg,
		log:           cfg.Logger,
		stop:          make(chan struct{}),
		inprocHandler: rpc.NewAdamniteRPCServer(),
		server:        &p2p.Server{Config: cfg.P2P},
	}

	node.rpcAPIs = append(node.rpcAPIs, node.apis()...)

	node.ipc = newIPCServer(node.log, cfg.IPCEndpoint())

	node.server.Config.Name = node.config.NodeName()
	node.server.Config.PrivateKey = node.config.NodeKey()
	node.server.Config.Logger = node.log

	if node.server.Config.NodeDatabase == "" {
		node.server.Config.NodeDatabase = node.config.NodeDB()
	}

	return node, nil
}

func (n *Node) Start() error {
	n.startStopLock.Lock()
	defer n.startStopLock.Unlock()

	n.lock.Lock()
	switch n.state {
	case runningState:
		n.lock.Unlock()
		return ErrNodeRunning
	case closedState:
		n.lock.Unlock()
		return ErrNodeStopped
	}
	n.state = runningState

	// Open networking and RPC endpoints
	err := n.openEndPoints()
	n.lock.Unlock()

	return err
}

func (n *Node) openEndPoints() error {
	n.log.Info("Starting P2P node", "instance", n.server.Name)
	if err := n.server.Start(); err != nil {
		return convertFileLockError(err)
	}
	err := n.startRPC()

	return err
}

func (n *Node) startRPC() error {
	if err := n.startInProc(); err != nil {
		return err
	}

	if n.ipc.endpoint != "" {
		if err := n.ipc.start(n.rpcAPIs); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) startInProc() error {
	for _, api := range n.rpcAPIs {
		if err := n.inprocHandler.RegisterName(api.Namespace, api.Service); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) Wait() {
	<-n.stop
}
