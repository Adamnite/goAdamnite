package rpc

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/rpc"
	"strings"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core"
	"github.com/vmihailenco/msgpack/v5"
)

type AdamniteServer struct {
	stateDB             *statedb.StateDB
	chain               *core.Blockchain
	addresses           []string
	GetContactsFunction DefaultGetContactsFunc
	listener            net.Listener
	Run                 func()
}

func (a *AdamniteServer) Addr() string {
	return a.listener.Addr().String()
}

type DefaultGetContactsFunc func() PassedContacts

func (a *AdamniteServer) Close() {
	_ = a.listener.Close()
}

const getContactsListEndpoint = "Adamnite.GetContactList"

func (a *AdamniteServer) GetContactList(params *[]byte, reply *[]byte) (err error) {
	contacts := a.GetContactsFunction()
	*reply, err = msgpack.Marshal(contacts)
	return
}

const getVersionEndpoint = "Adamnite.GetVersion"

func (a *AdamniteServer) GetVersion(params *[]byte, reply *AdmVersionReply) error {
	log.Println("[Adamnite RPC] Get Version")
	//TODO: add the versioning, of the blockchain, and have it passed here.
	//TODO: parse the parameters from this
	reply.Client_version = ""
	reply.Timestamp = time.Now().UTC()
	reply.Addr_received = ""
	reply.Addr_from = ""
	reply.Last_round = &big.Int{}

	return nil
}

const getChainIDEndpoint = "Adamnite.GetChainID"

func (a *AdamniteServer) GetChainID(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Get chain ID")
	if a.chain == nil || a.chain.Config() == nil {
		return errors.New("chain is not set")
	}

	data, err := msgpack.Marshal(a.chain.Config().ChainID.String())
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

const getBalanceEndpoint = "Adamnite.GetBalance"

func (a *AdamniteServer) GetBalance(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Get balance")
	input := struct {
		Address string
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	data, err := msgpack.Marshal(a.stateDB.GetBalance(common.HexToAddress(input.Address)).String())
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

const getAccountsEndpoint = "Adamnite.GetAccounts"

func (a *AdamniteServer) GetAccounts(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Get accounts")

	data, err := msgpack.Marshal(a.addresses)
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

const getBlockByHashEndpoint = "Adamnite.GetBlockByHash"

func (a *AdamniteServer) GetBlockByHash(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Get block by hash")

	input := struct {
		BlockHash common.Hash
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	data, err := msgpack.Marshal(*a.chain.GetBlockByHash(input.BlockHash))
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil

}

const getBlockByNumberEndpoint = "Adamnite.GetBlockByNumber"

func (a *AdamniteServer) GetBlockByNumber(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Get block by number")

	input := struct {
		BlockNumber big.Int
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	data, err := msgpack.Marshal(*a.chain.GetBlockByNumber(&input.BlockNumber))
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

const createAccountEndpoint = "Adamnite.CreateAccount"

func (a *AdamniteServer) CreateAccount(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Create account")

	input := struct {
		Address string
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	for _, address := range a.addresses {
		if address == input.Address {
			log.Println("[Adamnite RPC server] Specified account already exists on chain")
			return errors.New("specified account already exists on chain")
		}
	}

	a.stateDB.CreateAccount(common.HexToAddress(input.Address))
	a.addresses = append(a.addresses, input.Address)

	data, err := msgpack.Marshal(true)
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

const sendTransactionEndpoint = "Adamnite.SendTransaction"

func (a *AdamniteServer) SendTransaction(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Send transaction")

	input := struct {
		Hash string
		Raw  string
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	// TODO: send transaction to blockchain node

	data, err := msgpack.Marshal(true)
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

func NewAdamniteServer(stateDB *statedb.StateDB, chain *core.Blockchain, port uint32) *AdamniteServer {
	rpcServer := rpc.NewServer()

	adamnite := new(AdamniteServer)
	adamnite.stateDB = stateDB
	adamnite.chain = chain

	if err := rpcServer.Register(adamnite); err != nil {
		log.Fatal(err)
	}

	listener, _ := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	log.Println("[Adamnite RPC server] Endpoint:", listener.Addr().String())

	runFunc := func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("[Listener accept]", err)
				return
			}

			go func(conn net.Conn) {
				defer func() {
					if err = conn.Close(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
						log.Println(err)
					}
				}()
				_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				rpcServer.ServeConn(conn)
			}(conn)
		}
	}
	adamnite.listener = listener
	adamnite.Run = runFunc
	return adamnite
}
