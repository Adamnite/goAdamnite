package node

import (
	"errors"
	"sync"

	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/rpc"
)

type ipcServer struct {
	log    log15.Logger
	port   uint32
	mu     sync.Mutex
	server *rpc.AdamniteServer
}

func newIPCServer(log log15.Logger, port uint32) *ipcServer {
	return &ipcServer{log: log, port: port}
}

func (ipc *ipcServer) start() error {
	ipc.mu.Lock()
	defer ipc.mu.Unlock()

	ipc.server = rpc.NewAdamniteServer(nil, nil, ipc.port)
	go ipc.server.Run()

	ipc.log.Info("IPC endpoint opened", "url", ipc.server.Addr())
	return nil
}

func (ipc *ipcServer) stop() error {
	ipc.mu.Lock()
	defer ipc.mu.Unlock()

	if ipc.server == nil {
		return errors.New("Uninitialized IPC server")
	}

	ipc.server.Close()

	ipc.server = nil
	ipc.log.Info("IPC endpoint closed", "url", ipc.server.Addr())
	return nil
}
