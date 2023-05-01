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
)

type Adamnite struct {
	stateDB   *statedb.StateDB
	chain     *core.Blockchain
	addresses []string
	listener  net.Listener
	Run       func()
}

func (a *Adamnite) Addr() string {
	return a.listener.Addr().String()
}

func (a *Adamnite) Close() {
	_ = a.listener.Close()
}

const getChainIDEndpoint = "Adamnite.GetChainID"

func (a *Adamnite) GetChainID(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Get chain ID")
	if a.chain == nil || a.chain.Config() == nil {
		return errors.New("chain is not set")
	}

	data, err := Encode(a.chain.Config().ChainID.String())
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

const getBalanceEndpoint = "Adamnite.GetBalance"

func (a *Adamnite) GetBalance(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Get balance")
	input := struct {
		Address string
	}{}

	if err := Decode(*params, &input); err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	data, err := Encode(a.stateDB.GetBalance(common.HexToAddress(input.Address)).String())
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

const getAccountsEndpoint = "Adamnite.GetAccounts"

func (a *Adamnite) GetAccounts(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Get accounts")

	data, err := Encode(a.addresses)
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

const getBlockByHashEndpoint = "Adamnite.GetBlockByHash"

func (a *Adamnite) GetBlockByHash(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Get block by hash")

	input := struct {
		BlockHash common.Hash
	}{}

	if err := Decode(*params, &input); err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	data, err := Encode(*a.chain.GetBlockByHash(input.BlockHash))
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil

}

const getBlockByNumberEndpoint = "Adamnite.GetBlockByNumber"

func (a *Adamnite) GetBlockByNumber(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Get block by number")

	input := struct {
		BlockNumber big.Int
	}{}

	if err := Decode(*params, &input); err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	data, err := Encode(*a.chain.GetBlockByNumber(&input.BlockNumber))
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

const createAccountEndpoint = "Adamnite.CreateAccount"

func (a *Adamnite) CreateAccount(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Create account")

	input := struct {
		Address string
	}{}

	if err := Decode(*params, &input); err != nil {
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

	data, err := Encode(true)
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

const sendTransactionEndpoint = "Adamnite.SendTransaction"

func (a *Adamnite) SendTransaction(params *[]byte, reply *[]byte) error {
	log.Println("[Adamnite RPC server] Send transaction")

	input := struct {
		Hash string
		Raw  string
	}{}

	if err := Decode(*params, &input); err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	// TODO: send transaction to blockchain node

	data, err := Encode(true)
	if err != nil {
		log.Printf("[Adamnite RPC server] Error: %s", err)
		return err
	}

	*reply = data
	return nil
}

func NewAdamniteServer(stateDB *statedb.StateDB, chain *core.Blockchain, port uint32) *Adamnite {
	rpcServer := rpc.NewServer()

	adamnite := new(Adamnite)
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
					if err = conn.Close(); err != nil && !strings.Contains(err.Error(), "Use of closed network connection") {
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
