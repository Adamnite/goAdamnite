package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"reflect"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/ugorji/go/codec"
)

// type Query struct {
// 	Data []byte
// }

type AdamniteServer struct {
	Endpoint string
	id       int
	statedb  *statedb.StateDB
	chain    *core.Blockchain
}

const adm_getBalance_endpoint = "AdamniteServer.GetBalance"

func (a *AdamniteServer) GetBalance(add common.Address, ans *BigIntRPC) error {
	//arg is passed, so we should sanitize it, but for now we will assume it is a correctly formatted hash
	fmt.Println("Starting get balance server side")
	if a.statedb == nil {
		return ErrStateNotSet
	}

	*ans = BigIntReplyFromBigInt(*a.statedb.GetBalance(add))
	return nil
}

const adm_getChainID_endpoint = "AdamniteServer.GetChainID"

func (a *AdamniteServer) GetChainID(_ interface{}, ans *BigIntRPC) error {
	fmt.Println("Starting get Chain ID server side")
	if a.chain == nil || a.chain.Config() == nil {
		return ErrChainNotSet
	}
	*ans = BigIntReplyFromBigInt(*a.chain.Config().ChainID)
	return nil
}

const adm_getBlockByHash_endpoint = "AdamniteServer.GetBlockByHash"

func (a *AdamniteServer) GetBlockByHash(hash common.Hash, ans *types.Block) error {
	*ans = *a.chain.GetBlockByHash(hash)
	return nil
}
func (a *AdamniteServer) GetBlockByNumber(blockIndex BigIntRPC, ans *types.Block) error {
	*ans = *a.chain.GetBlockByNumber(blockIndex.toBigInt())
	return nil
}

func NewAdamniteServer(db *statedb.StateDB, chainReference *core.Blockchain) *AdamniteServer {
	admServer := new(AdamniteServer)
	admServer.statedb = db
	admServer.chain = chainReference
	return admServer
}

func (as *AdamniteServer) Launch() {
	// Start listening for the requests on any open port
	listener, err := net.Listen("tcp", "[127.0.0.1]:0")
	as.Endpoint = listener.Addr().String()
	// fmt.Println(as.Endpoint)
	// http.Serve(listener, nil)
	handler := rpc.NewServer()
	// handler.HandleHTTP("/", "/debug/")
	// handler.ServeCodec()
	handler.Register(as)
	if err != nil {
		fmt.Printf("listen(%q): %s\n", as.Endpoint, err)
		return
	}
	// fmt.Printf("Server %d listening on %s\n", as.id, listener.Addr())
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	go func() {
		for {
			cxn, err := listener.Accept()

			// handler.ServeCodec(codec.GoRpc.ServerCodec(cxn, handler))
			// codec.MsgpackSpecRpc.ServerCodec(cxn, &mh)
			rpcCodec := codec.GoRpc.ServerCodec(cxn, &mh)

			handler.ServeCodec(rpcCodec)
			if err != nil {
				log.Printf("listen(%q): %s\n", as.Endpoint, err)
				return
			}
			log.Printf("Server %d accepted connection to %s from %s\n", as.id, cxn.LocalAddr(), cxn.RemoteAddr())
			go handler.ServeConn(cxn)
		}
	}()
	fmt.Println("server launched!")
}
