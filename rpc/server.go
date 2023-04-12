package rpc

import (
	"encoding/base64"
	"errors"
	"log"
	"net"
	"net/rpc"
	"strings"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/vmihailenco/msgpack/v5"
)

type AdamniteServer struct {
	stateDB   *statedb.StateDB
	chain     *core.Blockchain
	addresses []string
}

func encodeBase64(value []byte) string {
	return base64.StdEncoding.EncodeToString(value)
}

const getChainIDEndpoint = "Adamnite.GetChainID"

func (a *AdamniteServer) GetChainID(params *[]byte, reply *string) error {
	log.Println("[Adamnite RPC] Get chain ID")
	if a.chain == nil || a.chain.Config() == nil {
		return errors.New("Chain is not set")
	}

	data, err := msgpack.Marshal(a.chain.Config().ChainID.String())
	if err != nil {
		return err
	}

	*reply = encodeBase64(data)
	return nil
}

const getBalanceEndpoint = "Adamnite.GetBalance"

func (a *AdamniteServer) GetBalance(params *[]byte, reply *string) error {
	log.Println("[Adamnite RPC] Get balance")
	input := struct {
		Address string
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		return err
	}

	data, err := msgpack.Marshal(a.stateDB.GetBalance(common.HexToAddress(input.Address)).String())
	if err != nil {
		return err
	}

	*reply = encodeBase64(data)
	return nil
}

const getAccountsEndpoint = "Adamnite.GetAccounts"

func (a *AdamniteServer) GetAccounts(params *[]byte, reply *string) error {
	log.Println("[Adamnite RPC] Get accounts")

	data, err := msgpack.Marshal(a.addresses)
	if err != nil {
		return err
	}

	*reply = encodeBase64(data)
	return nil
}

const getBlockByHashEndpoint = "Adamnite.GetBlockByHash"

func (a *AdamniteServer) GetBlockByHash(hash common.Hash, reply *types.Block) error {
	*reply = *a.chain.GetBlockByHash(hash)
	return nil
}

const getBlockByNumberEndpoint = "Adamnite.GetBlockByNumber"

func (a *AdamniteServer) GetBlockByNumber(blockIndex BigIntRPC, reply *types.Block) error {
	*reply = *a.chain.GetBlockByNumber(blockIndex.toBigInt())
	return nil
}

const createAccountEndpoint = "Adamnite.CreateAccount"

func (a *AdamniteServer) CreateAccount(params *[]byte, reply *string) error {
	log.Println("[Adamnite RPC] Create account")

	input := struct {
		Address string
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		return err
	}

	for _, address := range a.addresses {
		if address == input.Address {
			log.Println("[Adamnite RPC] Specified account already exists on chain")
			return errors.New("Specified account already exists on chain")
		}
	}

	a.stateDB.CreateAccount(common.HexToAddress(input.Address))
	a.addresses = append(a.addresses, input.Address)

	data, err := msgpack.Marshal(true)
	if err != nil {
		return err
	}

	*reply = encodeBase64(data)
	return nil
}

func NewAdamniteServer(stateDB *statedb.StateDB, chain *core.Blockchain) (listener net.Listener, runFunc func()) {
	adamnite := newAdamniteServerSetup(stateDB, chain)

	listener, runFunc, err := adamnite.setupServerListenerAndRunFuncs("127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	return listener, runFunc
}

// for internal use, to help the server be setup to handle more factors.
func newAdamniteServerSetup(stateDB *statedb.StateDB, chain *core.Blockchain) *AdamniteServer {
	adamnite := new(AdamniteServer)
	adamnite.stateDB = stateDB
	adamnite.chain = chain

	return adamnite
}

// for internal use to spin-up a server with a standard listener and runtime func
func (admServ *AdamniteServer) setupServerListenerAndRunFuncs(listenerPoint string) (net.Listener, func(), error) {
	rpcServer := rpc.NewServer()
	if err := rpcServer.Register(admServ); err != nil {
		log.Fatal(err)
		return nil, nil, err
	}

	listener, _ := net.Listen("tcp", listenerPoint)
	log.Println("[Adamnite RPC] Endpoint:", listener.Addr().String())

	runFunc := func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("[Listener accept]", err)
				return
			}

			go func(conn net.Conn) {
				defer func() {
					if err = conn.Close(); err != nil && !strings.Contains(err.Error(), "Use of closed network connection") {
						log.Println(err)
					}
				}()
				_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				rpcServer.ServeConn(conn)
			}(conn)
		}
	}
	return listener, runFunc, nil
}
