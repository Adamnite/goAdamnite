package node

import (
	"errors"
	"sync"

	"github.com/adamnite/go-adamnite/rpc"

	log "github.com/sirupsen/logrus"
)

type ipcServer struct {
	port   uint32
	mu     sync.Mutex
	server *rpc.AdamniteServer
}

func newIPCServer(port uint32) *ipcServer {
	return &ipcServer{port: port}
}

func (ipc *ipcServer) start() error {
	ipc.mu.Lock()
	defer ipc.mu.Unlock()

	ipc.server = rpc.NewAdamniteServer(ipc.port)
	go ipc.server.Run()

	log.Info("IPC endpoint opened", "url", ipc.server.Addr())
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
	log.Info("IPC endpoint closed", "url", ipc.server.Addr())
	return nil
}
