package main

import (
	"encoding/json"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"

	"github.com/adamnite/go-adamnite/adm/database"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/params"
	admRpc "github.com/adamnite/go-adamnite/rpc"
	"github.com/rs/cors"
)

// Test setup
var (
	niteBigExponent = big.NewInt(1).Exp(big.NewInt(10), big.NewInt(20), nil) // big math version of 10**20
	testAccounts    = []common.Address{
		common.BytesToAddress([]byte{0x00}),
		common.BytesToAddress([]byte{0x01}),
		common.BytesToAddress([]byte{0x02}),
	}
	testBalances = []*big.Int{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(2),
	}
)

type RPCRequest struct {
	Method string
	Params []byte
}

var RPCServerAddr *string

func main() {
	// Setup test blockchain
	db := rawdb.NewMemoryDB()
	stateDB, _ := statedb.New(common.Hash{}, statedb.NewDatabase(db))
	chainConfig := params.TestnetChainConfig

	for i, address := range testBalances {
		testBalances[i] = big.NewInt(0).Mul(niteBigExponent, address)
		stateDB.AddBalance(testAccounts[i], testBalances[i])
	}

	rootHash := stateDB.IntermediateRoot(false)
	stateDB.Database().TrieDB().Commit(rootHash, false, nil)

	blockchain, err := blockchain.NewBlockchain(
		db,
		chainConfig,
	)
	if err != nil {
		log.Println(err)
	}

	// Create RPC and HTTP servers
	bouncerServer := admRpc.NewBouncerServer(stateDB, blockchain, 0)
	defer func() {
		bouncerServer.Close()
	}()
	adamniteServer.Start()

	RPCServerAddr = new(string)
	*RPCServerAddr = bouncerServer.Addr()
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

		conn, err := rpc.Dial("tcp", *RPCServerAddr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() {
			_ = conn.Close()
		}()

		var reply []byte
		if err = conn.Call(req.Method, &req.Params, &reply); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Handle response
		w.Header().Set("Content-Type", "application/x-msgpack")
		w.Write(reply)
	})

	handler := cors.Default().Handler(mux)
	server := http.Server{Addr: "127.0.0.1:3000", Handler: handler}
	listener, _ := net.Listen("tcp", server.Addr)
	log.Println("[Adamnite HTTP] Endpoint:", listener.Addr().String()+"/v1/")

	_ = server.Serve(listener)
}
