package rpc

import (
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
	stateDB         *statedb.StateDB
	chain           *core.Blockchain
	hostingNodeID   common.Address
	seenConnections map[common.Hash]common.Void

	addresses                 []string
	GetContactsFunction       func() PassedContacts
	listener                  net.Listener
	mostRecentReceivedIP      string //TODO: CHECK THIS! Most likely can cause a race condition.
	timesTestHasBeenCalled    int
	newConnection             func(string, common.Address)
	forwardingMessageReceived func(ForwardingContent, []byte) error
	Run                       func()
}

func (a *AdamniteServer) Addr() string {
	return a.listener.Addr().String()
}
func (a *AdamniteServer) SetForwardFunc(newForward func(ForwardingContent, []byte) error) {
	a.forwardingMessageReceived = newForward
}
func (a *AdamniteServer) SetNewConnectionFunc(newConn func(string, common.Address)) {
	a.newConnection = newConn
}
func (a *AdamniteServer) SetHostingID(id *common.Address) {
	if id == nil {
		a.hostingNodeID = common.Address{0}
		return
	}
	a.hostingNodeID = *id
}
func (a *AdamniteServer) Close() {
	_ = a.listener.Close()
}

const serverPreface = "[Adamnite RPC server] %v \n"

const TestServerEndpoint = "AdamniteServer.TestServer"

func (a *AdamniteServer) TestServer(params *[]byte, reply *[]byte) error {
	log.Printf(serverPreface, "Test Server")
	a.timesTestHasBeenCalled++
	return nil
}
func (a *AdamniteServer) GetTestsCount() int {
	return a.timesTestHasBeenCalled
}

const forwardMessageEndpoint = "AdamniteServer.ForwardMessage"

func (a *AdamniteServer) ForwardMessage(params *[]byte, reply *[]byte) error {
	log.Printf(serverPreface, "Forward Message")
	if a.forwardingMessageReceived == nil {
		return ErrNotSetupToHandleForwarding
	}

	var content ForwardingContent
	if err := msgpack.Unmarshal(*params, &content); err != nil {
		return err
	}
	if _, exists := a.seenConnections[content.Signature]; exists {
		return ErrAlreadyForwarded //we've already seen it
	} else {
		a.seenConnections[content.Signature] = common.Void{} //this doesn't actually use any memory
	}

	if content.DestinationNode != nil && *content.DestinationNode != a.hostingNodeID {
		//it's not for us, but it is for someone!
		return a.forwardingMessageReceived(content, *reply)
	}

	if content.DestinationNode == nil {
		//its for everyone, so we need to parse it and share it
		if err := a.forwardingMessageReceived(content, *reply); err != nil {
			return err
		}
	}
	//we need to call it on ourselves.
	switch content.FinalEndpoint {
	case getContactsListEndpoint:
		return a.GetContactList(&content.FinalParams, &content.FinalReply)
	case TestServerEndpoint:
		return a.TestServer(&content.FinalParams, &content.FinalReply)
	}
	return nil
}

const getContactsListEndpoint = "AdamniteServer.GetContactList"

func (a *AdamniteServer) GetContactList(params *[]byte, reply *[]byte) (err error) {
	log.Printf(serverPreface, "Get Contact list")
	contacts := a.GetContactsFunction()
	*reply, err = msgpack.Marshal(contacts)
	return
}

const getVersionEndpoint = "AdamniteServer.GetVersion"

func (a *AdamniteServer) GetVersion(params *[]byte, reply *[]byte) error {
	log.Printf(serverPreface, "Get Version")
	// var receivedAddress common.Address //TODO, have the connection string passed as well.
	receivedData := struct {
		Address           common.Address
		HostingServerPort string
	}{}
	if err := msgpack.Unmarshal(*params, &receivedData); err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}
	if a.newConnection != nil && a.mostRecentReceivedIP != "" && receivedData.HostingServerPort != "" {
		//format it correctly to use the right port
		foo := strings.Split(a.mostRecentReceivedIP, ":")
		a.newConnection(fmt.Sprintf("%v:%v", foo[0], receivedData.HostingServerPort), receivedData.Address)
	}
	// if a.chain == nil {//leave this commented out until RPC consistently has a chain passed to it
	// 	return ErrChainNotSet
	// }
	ans := AdmVersionReply{}
	// ans.Client_version = a.chain.Config().ChainID.String() //TODO: replace this with a better versioning system
	ans.Timestamp = time.Now().UTC()
	ans.Addr_received = receivedData.Address
	ans.Addr_from = a.hostingNodeID //TODO: pass the hosting address down to the RPC
	// ans.Last_round = a.chain.CurrentBlock().Number()
	if data, err := msgpack.Marshal(ans); err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	} else {
		*reply = data
	}
	return nil
}

const getChainIDEndpoint = "AdamniteServer.GetChainID"

func (a *AdamniteServer) GetChainID(params *[]byte, reply *[]byte) error {
	log.Printf(serverPreface, "Get chain ID")
	if a.chain == nil || a.chain.Config() == nil {
		return ErrChainNotSet
	}

	data, err := msgpack.Marshal(a.chain.Config().ChainID.String())
	if err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}

	*reply = data
	return nil
}

const getBalanceEndpoint = "AdamniteServer.GetBalance"

func (a *AdamniteServer) GetBalance(params *[]byte, reply *[]byte) error {
	log.Printf(serverPreface, "Get balance")
	input := struct {
		Address string
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}

	data, err := msgpack.Marshal(a.stateDB.GetBalance(common.HexToAddress(input.Address)).String())
	if err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}

	*reply = data
	return nil
}

const getAccountsEndpoint = "AdamniteServer.GetAccounts"

func (a *AdamniteServer) GetAccounts(params *[]byte, reply *[]byte) error {
	log.Printf(serverPreface, "Get accounts")

	data, err := msgpack.Marshal(a.addresses)
	if err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}

	*reply = data
	return nil
}

const getBlockByHashEndpoint = "AdamniteServer.GetBlockByHash"

func (a *AdamniteServer) GetBlockByHash(params *[]byte, reply *[]byte) error {
	log.Printf(serverPreface, "Get block by hash")

	input := struct {
		BlockHash common.Hash
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}

	data, err := msgpack.Marshal(*a.chain.GetBlockByHash(input.BlockHash))
	if err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}

	*reply = data
	return nil

}

const getBlockByNumberEndpoint = "AdamniteServer.GetBlockByNumber"

func (a *AdamniteServer) GetBlockByNumber(params *[]byte, reply *[]byte) error {
	log.Printf(serverPreface, "Get block by number")

	input := struct {
		BlockNumber big.Int
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}

	data, err := msgpack.Marshal(*a.chain.GetBlockByNumber(&input.BlockNumber))
	if err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}

	*reply = data
	return nil
}

const createAccountEndpoint = "AdamniteServer.CreateAccount"

func (a *AdamniteServer) CreateAccount(params *[]byte, reply *[]byte) error {
	log.Printf(serverPreface, "Create account")

	input := struct {
		Address string
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}

	for _, address := range a.addresses {
		if address == input.Address {
			log.Printf(serverPreface, ErrPreExistingAccount)
			return ErrPreExistingAccount
		}
	}

	a.stateDB.CreateAccount(common.HexToAddress(input.Address))
	a.addresses = append(a.addresses, input.Address)

	data, err := msgpack.Marshal(true)
	if err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}

	*reply = data
	return nil
}

const sendTransactionEndpoint = "AdamniteServer.SendTransaction"

func (a *AdamniteServer) SendTransaction(params *[]byte, reply *[]byte) error {
	log.Printf(serverPreface, "Send transaction")

	input := struct {
		Hash string
		Raw  string
	}{}

	if err := msgpack.Unmarshal(*params, &input); err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		return err
	}

	// TODO: send transaction to blockchain node

	data, err := msgpack.Marshal(true)
	if err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
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
	adamnite.timesTestHasBeenCalled = 0
	adamnite.seenConnections = make(map[common.Hash]common.Void)

	if err := rpcServer.Register(adamnite); err != nil {
		log.Fatal(err)
	}

	listener, _ := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	log.Printf(serverPreface, fmt.Sprint("Endpoint: ", listener.Addr().String()))

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
				adamnite.mostRecentReceivedIP = conn.RemoteAddr().String()
				_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				rpcServer.ServeConn(conn)
			}(conn)
		}
	}
	adamnite.listener = listener
	adamnite.Run = runFunc
	return adamnite
}
