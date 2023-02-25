package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/vmihailenco/msgpack/v5"
)


type AdamniteServer struct {
	Endpoint string
	id       int
	statedb  statedb.StateDB
}

func (a *AdamniteServer) GetBalance(arg []byte, ans *Reply) error {
	//arg is passed, so we should sanitize it, but for now we will assume it is a correctly formatted hash
	var add common.Address
	err := msgpack.Unmarshal(arg, &add)
	if err != nil {
		fmt.Println(err)
		return err
	}
	balBytes, err := msgpack.Marshal(a.statedb.GetBalance(add))
	if err != nil {
		fmt.Println(err)
		return err
	}

	ans.Data = balBytes

	return nil
}

func NewAdamniteServer(db statedb.StateDB) *AdamniteServer {
	admServer := new(AdamniteServer)
	admServer.statedb = db
	return admServer
}

func (as *AdamniteServer) Launch() {
	// Start listening for the requests on any open port
	listener, err := net.Listen("tcp", "[127.0.0.1]:0")
	as.Endpoint = listener.Addr().String()
	// fmt.Println(as.Endpoint)
	// http.Serve(listener, nil)
	handler := rpc.NewServer()
	handler.HandleHTTP("/", "/debug/")
	handler.Register(as)
	if err != nil {
		fmt.Printf("listen(%q): %s\n", as.Endpoint, err)
		return
	}
	// fmt.Printf("Server %d listening on %s\n", as.id, listener.Addr())
	go func() {
		for {
			cxn, err := listener.Accept()
			if err != nil {
				log.Printf("listen(%q): %s\n", as.Endpoint, err)
				return
			}
			log.Printf("Server %d accepted connection to %s from %s\n", as.id, cxn.LocalAddr(), cxn.RemoteAddr())
			go handler.ServeConn(cxn)
		}
	}()
}
