package node

import (
	"errors"
	"path/filepath"
	"sync"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/event"
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

	eventmux *event.TypeMux

	startStopLock sync.Mutex
	state         int
	lock          sync.Mutex

	inprocHandler *rpc.Server
	rpcAPIs       []rpc.API
	services      []Service

	openedDatabases map[*OpenedDB]struct{}
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
		config:          cfg,
		log:             cfg.Logger,
		stop:            make(chan struct{}),
		inprocHandler:   rpc.NewAdamniteRPCServer(),
		server:          &p2p.Server{Config: cfg.P2P},
		openedDatabases: make(map[*OpenedDB]struct{}),
		eventmux:        new(event.TypeMux),
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

	services := make([]Service, len(n.services))
	copy(services, n.services)
	n.lock.Unlock()

	if err != nil {
		// n.doClose(nil)
		return err
	}

	// Start all registered services.
	var started []Service
	for _, service := range services {
		if err = service.Start(); err != nil {
			break
		}
		started = append(started, service)
	}
	// Check if any service failed to start.
	if err != nil {
		// n.stopServices(started)
		// n.doClose(nil)
	}
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

// ResolvePath returns the absolute path of a resource in the instance directory.
func (n *Node) ResolvePath(x string) string {
	return n.config.ResolvePath(x)
}

func (n *Node) OpenDatabase(fileName string, cache int, handle int, readonly bool) (adamnitedb.Database, error) {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.state == closedState {
		return nil, ErrNodeStopped
	}

	var db adamnitedb.Database
	var err error

	if n.config.DataDir == "" {
		return nil, errors.New("datadir directory does not exists")
	} else {
		dbPath := n.ResolvePath(fileName)
		db, err = rawdb.NewAdamniteLevelDB(dbPath, cache, handle, readonly)
	}

	if err == nil {
		db = n.wrapDatabase(db)
	}
	return db, nil
}

func (n *Node) Wait() {
	<-n.stop
}

type OpenedDB struct {
	adamnitedb.Database
	n *Node
}

func (db *OpenedDB) Close() error {
	db.n.lock.Lock()
	delete(db.n.openedDatabases, db)
	db.n.lock.Unlock()
	return db.Database.Close()
}

func (n *Node) closeAllDatabases() (errors []error) {
	for db := range n.openedDatabases {
		delete(n.openedDatabases, db)
		if err := db.Database.Close(); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (n *Node) wrapDatabase(db adamnitedb.Database) adamnitedb.Database {
	wrapper := &OpenedDB{db, n}
	n.openedDatabases[wrapper] = struct{}{}
	return wrapper
}

func (n *Node) Server() *p2p.Server {
	n.lock.Lock()
	defer n.lock.Unlock()

	return n.server
}

func (n *Node) RegistProtocols(protocols []p2p.Protocol) {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.state != initializingState {
		panic("cannot regist protocols on running or stopped node")
	}

	n.server.Protocols = append(n.server.Protocols, protocols...)
}

func (n *Node) RegistServices(serivce Service) {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.state != initializingState {
		panic("cannot regist service on runing or stopped node")
	}

	if containsService(n.services, serivce) {
		panic("The service %T was already registered")
	}

	n.services = append(n.services, serivce)
}

func (n *Node) EventMux() *event.TypeMux {
	return n.eventmux
}

func containsService(lfs []Service, l Service) bool {
	for _, obj := range lfs {
		if obj == l {
			return true
		}
	}
	return false
}
