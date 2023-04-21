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

type Adamnite struct {
	stateDB   *statedb.StateDB
	chain     *core.Blockchain
	addresses []string
}

func encodeBase64(value []byte) string {
	return base64.StdEncoding.EncodeToString(value)
}

const getChainIDEndpoint = "Adamnite.GetChainID"

func (a *Adamnite) GetChainID(params *[]byte, reply *string) error {
	log.Println("[Adamnite RPC] Get chain ID")
	if a.chain == nil || a.chain.Config() == nil {
		return errors.New("chain is not set")
	}

	data, err := msgpack.Marshal(a.chain.Config().ChainID.String())
	if err != nil {
		log.Printf("[Adamnite RPC] Error: %s", err)
		return err
	}

	*reply = encodeBase64(data)
	return nil
}

const getBalanceEndpoint = "Adamnite.GetBalance"

func (a *Adamnite) GetBalance(params *[]byte, reply *string) error {
	log.Println("[Adamnite RPC] Get balance")
	input := struct {
		Address string
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf("[Adamnite RPC] Error: %s", err)
		return err
	}

	data, err := msgpack.Marshal(a.stateDB.GetBalance(common.HexToAddress(input.Address)).String())
	if err != nil {
		log.Printf("[Adamnite RPC] Error: %s", err)
		return err
	}

	*reply = encodeBase64(data)
	return nil
}

const getAccountsEndpoint = "Adamnite.GetAccounts"

func (a *Adamnite) GetAccounts(params *[]byte, reply *string) error {
	log.Println("[Adamnite RPC] Get accounts")

	data, err := msgpack.Marshal(a.addresses)
	if err != nil {
		log.Printf("[Adamnite RPC] Error: %s", err)
		return err
	}

	*reply = encodeBase64(data)
	return nil
}

const getBlockByHashEndpoint = "Adamnite.GetBlockByHash"

func (a *Adamnite) GetBlockByHash(hash common.Hash, reply *types.Block) error {
	*reply = *a.chain.GetBlockByHash(hash)
	return nil
}

const getBlockByNumberEndpoint = "Adamnite.GetBlockByNumber"

func (a *Adamnite) GetBlockByNumber(blockIndex BigIntRPC, reply *types.Block) error {
	*reply = *a.chain.GetBlockByNumber(blockIndex.toBigInt())
	return nil
}

const createAccountEndpoint = "Adamnite.CreateAccount"

func (a *Adamnite) CreateAccount(params *[]byte, reply *string) error {
	log.Println("[Adamnite RPC] Create account")

	input := struct {
		Address string
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf("[Adamnite RPC] Error: %s", err)
		return err
	}

	for _, address := range a.addresses {
		if address == input.Address {
			log.Println("[Adamnite RPC] Specified account already exists on chain")
			return errors.New("specified account already exists on chain")
		}
	}

	a.stateDB.CreateAccount(common.HexToAddress(input.Address))
	a.addresses = append(a.addresses, input.Address)

	data, err := msgpack.Marshal(true)
	if err != nil {
		log.Printf("[Adamnite RPC] Error: %s", err)
		return err
	}

	*reply = encodeBase64(data)
	return nil
}

const sendTransactionEndpoint = "Adamnite.SendTransaction"

func (a *Adamnite) SendTransaction(params *[]byte, reply *string) error {
	log.Println("[Adamnite RPC] Send transaction")

	input := struct {
		Hash string
		Raw  string
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf("[Adamnite RPC] Error: %s", err)
		return err
	}

	// TODO: send transaction to blockchain node

	data, err := msgpack.Marshal(true)
	if err != nil {
		log.Printf("[Adamnite RPC] Error: %s", err)
		return err
	}

	*reply = encodeBase64(data)
	return nil
}

func NewAdamniteServer(stateDB *statedb.StateDB, chain *core.Blockchain) (listener net.Listener, runFunc func()) {
	rpcServer := rpc.NewServer()

	adamnite := new(Adamnite)
	adamnite.stateDB = stateDB
	adamnite.chain = chain

	if err := rpcServer.Register(adamnite); err != nil {
		log.Fatal(err)
	}

	listener, _ = net.Listen("tcp", "127.0.0.1:0")
	log.Println("[Adamnite RPC] Endpoint:", listener.Addr().String())

	runFunc = func() {
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
	return listener, runFunc
}
