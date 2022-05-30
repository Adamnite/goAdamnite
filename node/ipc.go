package node

import (
	"net"
	"sync"

	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/rpc"
)

type ipcServer struct {
	log      log15.Logger
	endpoint string

	listener net.Listener
	mu       sync.Mutex
	server   *rpc.Server
}

func newIPCServer(log log15.Logger, endpoint string) *ipcServer {
	return &ipcServer{log: log, endpoint: endpoint}
}

func (ipc *ipcServer) start(apis []rpc.API) error {
	ipc.mu.Lock()
	defer ipc.mu.Unlock()

	if ipc.listener != nil {
		return nil
	}

	listener, srv, err := rpc.StartAdamniteIPCEndpoint(ipc.endpoint, apis)
	if err != nil {
		ipc.log.Warn("IPC opening failed", "url", ipc.endpoint, "error", err)
		return err
	}

	ipc.log.Info("IPC endpoint opened", "url", ipc.endpoint)
	ipc.server, ipc.listener = srv, listener
	return nil
}

func (ipc *ipcServer) stop() error {
	ipc.mu.Lock()
	defer ipc.mu.Unlock()

	if ipc.listener == nil {
		return nil
	}

	err := ipc.listener.Close()
	ipc.server.Stop()
	ipc.listener, ipc.server = nil, nil
	ipc.log.Info("IPC endpoint closed", "url", ipc.endpoint)
	return err
}
