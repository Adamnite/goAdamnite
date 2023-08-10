package node

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/bargossip"
	"github.com/adamnite/go-adamnite/event"
	"github.com/adamnite/go-adamnite/rpc"

	log "github.com/sirupsen/logrus"
)

const (
	initializingState = iota
	runningState
	closedState
)

type Node struct {
	config *Config
	stop   chan struct{}

	ipc    *ipcServer
	server *bargossip.Server

	eventmux *event.TypeMux

	startStopLock sync.Mutex
	state         int
	lock          sync.Mutex

	adamniteServer *rpc.AdamniteServer
	rpcAPIs        []rpc.AdamniteServer
	services       []Service

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

	node := &Node{
		config:          cfg,
		stop:            make(chan struct{}),
		adamniteServer:  rpc.NewAdamniteServer(0),
		server:          &bargossip.Server{Config: cfg.P2P},
		openedDatabases: make(map[*OpenedDB]struct{}),
		eventmux:        new(event.TypeMux),
	}

	node.ipc = newIPCServer(0)

	node.server.Config.Name = node.config.NodeName()
	node.server.Config.ServerPrvKey = node.config.NodeKey()

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
		n.releaseResources(nil)
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
		n.stopServices(started)
		n.releaseResources(nil)
	}
	return err
}

func (n *Node) openEndPoints() error {
	log.Info("Starting P2P node", "instance", n.server.Name)
	if err := n.server.Start(); err != nil {
		return convertFileLockError(err)
	}
	err := n.startRPC()
	if err != nil {
		n.stopRPC()
		n.server.Stop()
	}

	return err
}

func (n *Node) startRPC() error {
	if err := n.startInProc(); err != nil {
		return err
	}

	if err := n.ipc.start(); err != nil {
		return err
	}

	return nil
}

func (n *Node) stopRPC() {
	n.ipc.stop()
}

func (n *Node) startInProc() error {
	// for _, api := range n.rpcAPIs {
	// 	if err := n.inprocHandler.RegisterName(api.Namespace, api.Service); err != nil {
	// 		return err
	// 	}
	// }
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
		db = rawdb.NewMemoryDB()
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

func (n *Node) Server() *bargossip.Server {
	n.lock.Lock()
	defer n.lock.Unlock()

	return n.server
}

func (n *Node) RegistProtocols(protocols []bargossip.SubProtocol) {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.state != initializingState {
		panic("cannot regist protocols on running or stopped node")
	}

	n.server.ChainProtocol = append(n.server.ChainProtocol, protocols...)
}

func (n *Node) RegistServices(serivce Service) {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.state != initializingState {
		panic("cannot register service on runing or stopped node")
	}

	if containsService(n.services, serivce) {
		panic(fmt.Sprintf("The service %T was already registered", serivce))
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

func (n *Node) Close() error {
	n.startStopLock.Lock()
	defer n.startStopLock.Unlock()

	n.lock.Lock()
	state := n.state
	n.lock.Unlock()

	switch state {
	case initializingState:
		return n.releaseResources(nil)
	case runningState:
		var errs []error
		if err := n.stopServices(n.services); err != nil {
			errs = append(errs, err)
		}
		return n.releaseResources(errs)
	case closedState:
		return ErrNodeStopped
	default:
		panic(fmt.Sprintf("Unknown node state: %d", state))
	}
}

func (n *Node) stopServices(running []Service) error {
	stopError := &NodeServiceStopError{Services: make(map[reflect.Type]error)}

	for i := len(running) - 1; i >= 0; i-- {
		if err := running[i].Stop(); err != nil {
			stopError.Services[reflect.TypeOf(running[i])] = err
		}
	}

	if len(stopError.Services) > 0 {
		return stopError
	}
	return nil
}

func (n *Node) releaseResources(errs []error) error {
	n.stopRPC()

	n.lock.Lock()
	n.state = closedState
	dbErrs := n.closeAllDatabases()
	errs = append(errs, dbErrs...)
	n.lock.Unlock()

	close(n.stop)
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[1]
	default:
		return fmt.Errorf("%v", errs)
	}
}
