package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strings"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"
	encoding "github.com/vmihailenco/msgpack/v5"
)

//bouncer acts as the endpoint handler for points primarily called by external clients (eg, those who weren't there when the data was passed, or need select data)

type BouncerServer struct {
	stateDB     *statedb.StateDB
	chain       *blockchain.Blockchain
	addresses   []string
	listener    net.Listener
	DebugOutput bool

	propagator  func(ForwardingContent, *[]byte) error
	getMessages func(common.Address, common.Address) []*utils.CaesarMessage
}

const bouncerPreface = "[Adamnite Bouncer RPC server] %v \n"

func (b *BouncerServer) print(methodName string) {
	if b.DebugOutput {
		log.Printf(bouncerPreface, methodName)
	}
}
func (b *BouncerServer) printError(methodName string, err error) {
	log.Printf(bouncerPreface, fmt.Sprintf("%v\tError: %s", methodName, err))
}

func (b *BouncerServer) SetHandlers(propagator func(ForwardingContent, *[]byte) error) {
	b.propagator = propagator
}
func (b *BouncerServer) SetMessagingHandlers(getMsg func(common.Address, common.Address) []*utils.CaesarMessage) {
	b.getMessages = getMsg
}

func NewBouncerServer(stateDB *statedb.StateDB, chain *blockchain.Blockchain, port uint32) *BouncerServer {
	rpcServer := rpc.NewServer()

	bouncer := new(BouncerServer)
	bouncer.stateDB = stateDB
	bouncer.chain = chain
	bouncer.DebugOutput = false
	bouncer.propagator = func(ForwardingContent, *[]byte) error {
		return fmt.Errorf("this is an incomplete bouncer server, and cannot forward")
	}

	if err := rpcServer.Register(bouncer); err != nil {
		log.Fatal(err)
	}

	listener, _ := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	log.Printf(bouncerPreface, fmt.Sprint("Bouncer Endpoint: ", listener.Addr().String()))
	bouncer.listener = listener
	go func() {
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
				rpcServer.ServeConn(conn)
			}(conn)
		}
	}()
	return bouncer
}
func (b *BouncerServer) Close() {
	//TODO: clear all mappings!
	_ = b.listener.Close()
}
func (b *BouncerServer) Addr() string {
	return b.listener.Addr().String()
}

const getBalanceEndpoint = "BouncerServer.GetBalance"

func (b *BouncerServer) GetBalance(params *[]byte, reply *[]byte) error {
	b.print("Get balance")
	var input string

	if err := encoding.Unmarshal(*params, &input); err != nil {
		b.printError("Get balance", err)
		return err
	}

	data, err := encoding.Marshal(b.stateDB.GetBalance(common.HexToAddress(input)).String())
	if err != nil {
		b.printError("Get balance", err)
		return err
	}

	*reply = data
	return nil
}

const getAccountsEndpoint = "BouncerServer.GetAccounts"

func (b *BouncerServer) GetAccounts(params *[]byte, reply *[]byte) error {
	b.print("Get accounts")

	data, err := encoding.Marshal(b.addresses)
	if err != nil {
		b.printError("Get accounts", err)
		return err
	}

	*reply = data
	return nil
}

const createAccountEndpoint = "BouncerServer.CreateAccount"

func (b *BouncerServer) CreateAccount(params *[]byte, reply *[]byte) error {
	b.print("Create account")

	var inputAddress string

	if err := encoding.Unmarshal(*params, &inputAddress); err != nil {
		b.printError("Create account", err)
		return err
	}

	for _, address := range b.addresses {
		if address == inputAddress {
			log.Printf(serverPreface, ErrPreExistingAccount)
			return ErrPreExistingAccount
		}
	}

	b.stateDB.CreateAccount(common.HexToAddress(inputAddress))
	b.addresses = append(b.addresses, inputAddress)

	data, err := encoding.Marshal(true)
	if err != nil {
		b.printError("Create account", err)
		return err
	}

	*reply = data
	return nil
}

const bouncerNewMessageEndpoint = "BouncerServer.NewMessage"

func (b *BouncerServer) NewMessage(params *[]byte, reply *[]byte) error {
	b.print("New Message")

	var msg utils.CaesarMessage

	if err := encoding.Unmarshal(*params, &msg); err != nil {
		b.printError("New Message", err)
		return err
	}
	if !msg.Verify() {
		return utils.ErrIncorrectSigner
	}
	if forwardableMsg, err := CreateForwardToAll(msg); err != nil {
		b.printError("New Message", err)
		return err
	} else {
		data, _ := encoding.Marshal(true)
		*reply = data
		return b.propagator(forwardableMsg, params)
	}
}

const bouncerGetMessagesBetween = "BouncerServer.GetMessagesBetween"

func (b *BouncerServer) GetMessagesBetween(params, reply *[]byte) error {
	b.print("get messages between")
	if b.getMessages == nil {
		return fmt.Errorf("not setup to return messages")
	}
	input := struct {
		A common.Address
		B common.Address
	}{}
	if err := encoding.Unmarshal(*params, &input); err != nil {
		b.printError("get messages between", err)
		return err
	}
	msgs := b.getMessages(input.A, input.B)

	ansBytes, err := encoding.Marshal(msgs)
	*reply = ansBytes
	return err
}
