package node

import (
	"sync"

	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/rpc"
)

type ipcServer struct {
	log      log15.Logger
	endpoint string
	mu       sync.Mutex
	server   *rpc.AdamniteServer
}

func newIPCServer(log log15.Logger, endpoint string) *ipcServer {
	return &ipcServer{log: log, endpoint: endpoint}
}

func (ipc *ipcServer) start() error {
	ipc.mu.Lock()
	defer ipc.mu.Unlock()

	srv := rpc.NewAdamniteServer(nil, nil)
	err := srv.Launch(&ipc.endpoint)

	if err != nil {
		ipc.log.Warn("IPC opening failed", "url", ipc.endpoint, "error", err)
		return err
	}

	ipc.log.Info("IPC endpoint opened", "url", ipc.endpoint)
	ipc.server = srv
	return nil
}

func (ipc *ipcServer) stop() error {
	ipc.mu.Lock()
	defer ipc.mu.Unlock()

	if ipc.server == nil {
		return nil
	}

	err := ipc.server.Stop()

	ipc.server = nil
	ipc.log.Info("IPC endpoint closed", "url", ipc.endpoint)
	return err
}
