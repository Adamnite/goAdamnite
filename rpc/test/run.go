package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/dpos"
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
	Params string
	Id     int
}

var RPCServerAddr *string

func decodeBase64(value *string) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(*value)))
	n, err := base64.StdEncoding.Decode(decoded, []byte(*value))
	if err != nil {
		return nil, err
	}
	return decoded[:n], nil
}

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

	blockchain, err := core.NewBlockchain(
		db,
		chainConfig,
		dpos.New(chainConfig, db),
	)
	if err != nil {
		log.Println(err)
	}

	// Create RPC and HTTP servers
	listenerRPC, rpcServerRunFunc, _ := admRpc.NewAdamniteServer(stateDB, blockchain)
	defer func() {
		_ = listenerRPC.Close()
	}()
	go rpcServerRunFunc()

	RPCServerAddr = new(string)
	*RPCServerAddr = listenerRPC.Addr().String()
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

		var reply string

		params, err := decodeBase64(&req.Params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = conn.Call(req.Method, params, &reply); err != nil {
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

	handler := cors.Default().Handler(mux)
	server := http.Server{Addr: "127.0.0.1:3000", Handler: handler}
	listener, _ := net.Listen("tcp", server.Addr)
	log.Println("[Adamnite HTTP] Endpoint:", listener.Addr().String()+"/v1/")

	_ = server.Serve(listener)
}
