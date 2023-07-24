package rpc

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/rpc"
	"strings"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/utils/bytes"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	encoding "github.com/vmihailenco/msgpack/v5"
)

//bouncer acts as the endpoint handler for points primarily called by external clients (eg, those who weren't there when the data was passed, or need select data)

type MessageKey struct {
	FromPublicKey string
	ToPublicKey   string
}

type BouncerServer struct {
	stateDB     *statedb.StateDB
	chain       *blockchain.Blockchain
	addresses   []string
	listener    net.Listener
	DebugOutput bool
	Version     string

	propagator  func(ForwardingContent, *[]byte) error
	getMessages func(bytes.Address, bytes.Address) []*utils.CaesarMessage

	messages map[*MessageKey][]*utils.CaesarMessage
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
func (b *BouncerServer) SetMessagingHandlers(getMsg func(bytes.Address, bytes.Address) []*utils.CaesarMessage) {
	b.getMessages = getMsg
}

func NewBouncerServer(stateDB *statedb.StateDB, chain *blockchain.Blockchain, port uint32) *BouncerServer {
	rpcServer := rpc.NewServer()

	bouncer := new(BouncerServer)
	bouncer.stateDB = stateDB
	bouncer.chain = chain
	bouncer.DebugOutput = false
	bouncer.Version = "0.1.2"
	bouncer.messages = make(map[*MessageKey][]*utils.CaesarMessage)
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

const getChainIDEndpoint = "BouncerServer.GetChainID"

func (b *BouncerServer) GetChainID(params *[]byte, reply *[]byte) error {
	b.print("Get chain ID")

	data, err := encoding.Marshal(b.Version)
	if err != nil {
		b.printError("Get chain ID", err)
		return err
	}

	*reply = data
	return nil
}

const getBalanceEndpoint = "BouncerServer.GetBalance"

func (b *BouncerServer) GetBalance(params *[]byte, reply *[]byte) error {
	b.print("Get balance")

	input := struct {
		Address string
	}{}

	if err := encoding.Unmarshal(*params, &input); err != nil {
		b.printError("Get balance", err)
		return err
	}

	data, err := encoding.Marshal(b.stateDB.GetBalance(bytes.HexToAddress(input.Address)).String())
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

	input := struct {
		Address string
	}{}

	if err := encoding.Unmarshal(*params, &input); err != nil {
		b.printError("Create account", err)
		return err
	}

	for _, address := range b.addresses {
		if address == input.Address {
			log.Printf(serverPreface, ErrPreExistingAccount)
			return ErrPreExistingAccount
		}
	}

	b.stateDB.CreateAccount(bytes.HexToAddress(input.Address))
	b.addresses = append(b.addresses, input.Address)

	data, err := encoding.Marshal(true)
	if err != nil {
		b.printError("Create account", err)
		return err
	}

	*reply = data
	return nil
}

const getBlockByHashEndpoint = "BouncerServer.GetBlockByHash"

func (b *BouncerServer) GetBlockByHash(params *[]byte, reply *[]byte) error {
	b.print("Get block by hash")

	input := struct {
		BlockHash bytes.Hash
	}{}

	if err := encoding.Unmarshal(*params, &input); err != nil {
		b.printError("Get block by hash", err)
		return err
	}

	data, err := encoding.Marshal(*b.chain.GetBlockByHash(input.BlockHash))
	if err != nil {
		b.printError("Get block by hash", err)
		return err
	}

	*reply = data
	return nil
}

const getBlockByNumberEndpoint = "BouncerServer.GetBlockByNumber"

func (b *BouncerServer) GetBlockByNumber(params *[]byte, reply *[]byte) error {
	b.print("Get block by number")

	input := struct {
		BlockNumber big.Int
	}{}

	if err := encoding.Unmarshal(*params, &input); err != nil {
		b.printError("Get block by number", err)
		return err
	}

	data, err := encoding.Marshal(*b.chain.GetBlockByNumber(&input.BlockNumber))
	if err != nil {
		b.printError("Get block by number", err)
		return err
	}

	*reply = data
	return nil
}

const bouncerNewMessageEndpoint = "BouncerServer.NewMessage"

func (b *BouncerServer) NewMessage(params *[]byte, reply *[]byte) error {
	b.print("New Message")

	input := struct {
		FromPublicKey string
		ToPublicKey   string
		RawMessage    string
		SignedMessage string
	}{}

	if err := encoding.Unmarshal(*params, &input); err != nil {
		b.printError("New Message", err)
		return err
	}

	msg := utils.NewSignedCaesarMessage(
		*accounts.AccountFromPubBytes(bytes.FromHex(input.ToPublicKey)),
		*accounts.AccountFromPubBytes(bytes.FromHex(input.FromPublicKey)),
		bytes.FromHex(input.RawMessage),
		bytes.FromHex(input.SignedMessage))

	// TODO: Verify the message

	k := &MessageKey{
		input.FromPublicKey,
		input.ToPublicKey,
	}
	b.messages[k] = append(b.messages[k], msg)

	data, _ := encoding.Marshal(true)
	*reply = data
	return nil
}

const bouncerGetMessages = "BouncerServer.GetMessages"

func (b *BouncerServer) GetMessages(params *[]byte, reply *[]byte) error {
	b.print("Get messages")

	input := &MessageKey{}

	if err := encoding.Unmarshal(*params, input); err != nil {
		b.printError("Get messages", err)
		return err
	}

	encryptedMessages := make(map[int64]string)
	for k, v := range b.messages {
		if k.FromPublicKey == input.FromPublicKey && k.ToPublicKey == input.ToPublicKey {
			for _, m := range v {
				encryptedMessages[m.InitialTime] = hex.EncodeToString(m.Message)
			}
		}
	}

	data, err := encoding.Marshal(encryptedMessages)
	if err != nil {
		b.printError("Get messages", err)
		return err
	}

	*reply = data
	return nil
}

const sendTransactionEndpoint = "BouncerServer.SendTransaction"

func (b *BouncerServer) SendTransaction(params *[]byte, reply *[]byte) error {
	b.print("Send transaction")

	var input utils.TransactionType

	if err := encoding.Unmarshal(*params, &input); err != nil {
		b.printError("Send transaction", err)
		return err
	}
	data, err := encoding.Marshal(true)
	if err != nil {
		b.printError("Send transaction", err)
		return err
	}
	*reply = data

	// if b.newTransactionReceived != nil {
	// 	// return b.newTransactionReceived(input, params)
	// } else {
	// 	//TODO: this node cant forward this transaction at all.
	// }

	return nil
}
