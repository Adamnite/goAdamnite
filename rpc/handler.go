package rpc

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/core"
)

type AdamNetworkHandler struct {
	tcpServer          *AdamniteServer
	tcpServersListener net.Listener
	httpHostConnection *AdamniteClient //this uses a client to forward requests to
	tcpConnections     []AdamniteClient
	//TODO: use the listening addresses as a way to also reference the tcp connections.
}

func newAdamniteNetworkHandler() *AdamNetworkHandler {
	anh := AdamNetworkHandler{
		tcpServer:      nil,
		tcpConnections: make([]AdamniteClient, 0),
	}
	return &anh
}

// for setting up a TCP adamnite server
func (anh *AdamNetworkHandler) SetupAdamniteServer(stateDB *statedb.StateDB, chain *core.Blockchain) {
	anh.tcpServer = newAdamniteServerSetup(stateDB, chain)
}

// For launching a pre established adamnite TCP server.
func (anh *AdamNetworkHandler) LaunchAdamniteServer(listenerPoint *string) error {
	if anh.tcpServer == nil {
		return fmt.Errorf("adamnite TCP server is not established to be launched")
	}
	var trueListeningPoint string
	if listenerPoint == nil { //basically if you don't want to set a port for this, pass a nil.
		trueListeningPoint = "127.0.0.1:0"
	} else {
		trueListeningPoint = *listenerPoint
	}
	tcpServersListener, startFunc, err := anh.tcpServer.setupServerListenerAndRunFuncs(trueListeningPoint)
	if err != nil {
		return err
	}
	anh.tcpServersListener = tcpServersListener
	defer func() {
		_ = tcpServersListener.Close()
	}()
	go startFunc()
	return nil
}

func (anh *AdamNetworkHandler) StopAdamniteServer() error {
	return anh.tcpServersListener.Close()
}
func (anh *AdamNetworkHandler) LaunchNewClientConnection(listenerPoint string) {
	adc := NewAdamniteClient(listenerPoint)
	anh.tcpConnections = append(anh.tcpConnections, *adc)
}

func (anh *AdamNetworkHandler) HandleHttpServer() error {
	var RPCServerAddr string
	//if we already have a server running locally, use that. Otherwise, use the first connection we already have going
	if anh.tcpServer != nil && anh.tcpServersListener != nil {
		RPCServerAddr = anh.tcpServersListener.Addr().String()
	} else if len(anh.tcpConnections) > 0 {
		RPCServerAddr = anh.tcpConnections[0].callAddress
	} else {
		//no server is running, and no connections are made, therefor, we cant host an HTTP server
		return fmt.Errorf("no available endpoints to relay HTTP through")
	}
	adamClientForHttpForwarding := NewAdamniteClient(RPCServerAddr)
	anh.httpHostConnection = adamClientForHttpForwarding

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Content-Type") != "application/x-msgpack" {
			http.Error(w, "Invalid Content-Type header value", http.StatusBadRequest)
			return
		}

		var req RPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer func() {
			adamClientForHttpForwarding.Close()
		}()

		var reply string

		params, err := decodeBase64(&req.Params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = adamClientForHttpForwarding.client.Call(req.Method, params, &reply); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Handle response
		w.Header().Set("Content-Type", "application/x-msgpack")
		resultBytes, _ := json.Marshal(struct {
			Message string
		}{
			reply,
		})
		if _, err = fmt.Fprintln(w, string(resultBytes)); err != nil {
			log.Println(err)
		}
	})
	return nil
}
