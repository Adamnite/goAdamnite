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
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/utils/bytes"
	"github.com/adamnite/go-adamnite/utils"
	encoding "github.com/vmihailenco/msgpack/v5"
)

type AdamniteServer struct {
	stateDB         *statedb.StateDB
	chain           *blockchain.Blockchain
<<<<<<< Updated upstream
	hostingNodeID   common.Address
	seenConnections map[common.Hash]common.Void
=======
	hostingNodeID   bytes.Address
	externalIP      string
	seenConnections syncmap.Map // map[bytes.Hash]bytes.Void
>>>>>>> Stashed changes

	addresses                 []string
	GetContactsFunction       func() PassedContacts
	listener                  net.Listener
	mostRecentReceivedIP      string //TODO: CHECK THIS! Most likely can cause a race condition.
	timesTestHasBeenCalled    int
	newConnection             func(string, bytes.Address)
	forwardingMessageReceived func(ForwardingContent, *[]byte) error
	newTransactionReceived    func(*utils.Transaction, *[]byte) error
	newCandidateHandler       func(utils.Candidate) error
	newVoteHandler            func(utils.Voter) error
	newMessageHandler         func(*utils.CaesarMessage)
	Run                       func()
	DebugOutput               bool
}

func (a *AdamniteServer) Addr() string {
	return a.listener.Addr().String()
}
func (a *AdamniteServer) SetHandlers(
	newForward func(ForwardingContent, *[]byte) error,
<<<<<<< Updated upstream
	newConn func(string, common.Address),
	newTransaction func(*utils.Transaction, *[]byte) error) {
=======
	newConn func(string, bytes.Address),
	newTransaction func(utils.TransactionType) error,
	newBlock func(utils.BlockType) error) {
>>>>>>> Stashed changes
	a.forwardingMessageReceived = newForward
	a.newConnection = newConn
	a.newTransactionReceived = newTransaction
}
func (a *AdamniteServer) SetForwardFunc(newForward func(ForwardingContent, *[]byte) error) {
	a.forwardingMessageReceived = newForward
}
func (a *AdamniteServer) SetNewConnectionFunc(newConn func(string, bytes.Address)) {
	a.newConnection = newConn
}

<<<<<<< Updated upstream
// set a response point if we get asked to handle a transaction
func (a *AdamniteServer) SetTransactionHandler(handler func(*utils.Transaction, *[]byte) error) {
	a.newTransactionReceived = handler
}

func (a *AdamniteServer) SetHostingID(id *common.Address) {
=======
func (a *AdamniteServer) SetHostingID(id *bytes.Address) {
>>>>>>> Stashed changes
	if id == nil {
		a.hostingNodeID = bytes.Address{0}
		return
	}
	a.hostingNodeID = *id
}
func (a *AdamniteServer) Close() {
	_ = a.listener.Close()
}

const serverPreface = "[Adamnite RPC server] %v \n"

func (a *AdamniteServer) print(methodName string) {
	if a.DebugOutput {
		log.Printf(serverPreface, methodName)
	}
}
func (a *AdamniteServer) printError(methodName string, err error) {
	log.Printf(serverPreface, fmt.Sprintf("%v\tError: %s", methodName, err))
}
func (a *AdamniteServer) AlreadySeen(fc ForwardingContent) bool {
	if _, exists := a.seenConnections[fc.Hash()]; exists {
		return true //we've already seen it
	}
	a.seenConnections[fc.Hash()] = common.Void{} //this doesn't actually use any memory
	return false
}

const TestServerEndpoint = "AdamniteServer.TestServer"

func (a *AdamniteServer) TestServer(params *[]byte, reply *[]byte) error {
	a.print("Test Server")
	a.timesTestHasBeenCalled++
	return nil
}
func (a *AdamniteServer) GetTestsCount() int {
	return a.timesTestHasBeenCalled
}

const forwardMessageEndpoint = "AdamniteServer.ForwardMessage"

func (a *AdamniteServer) ForwardMessage(params *[]byte, reply *[]byte) error {
	a.print("Forward Message")
	if a.forwardingMessageReceived == nil {
		return ErrNotSetupToHandleForwarding
	}

	var content ForwardingContent
	if err := encoding.Unmarshal(*params, &content); err != nil {
		return err
	}
	if a.AlreadySeen(content) {
		return ErrAlreadyForwarded
	}
	if content.DestinationNode != nil { //targeted
		if *content.DestinationNode == a.hostingNodeID {
			//its been relayed directly to us
			log.Println("being called directly on us")
			return a.callOnSelf(content)
		}
		if *content.DestinationNode != a.hostingNodeID {
			log.Println("being called directly on someone")
			//it's not for us, but it is for someone!
			return a.forwardingMessageReceived(content, reply)
		}
	}
	//for everyone, including us
	return a.callOnSelfThenShare(content)
}

// directly handle the message on ourselves.
func (a *AdamniteServer) callOnSelf(content ForwardingContent) error {
	log.Println("call on self")
	switch content.FinalEndpoint {
	case NewCandidateEndpoint:
		return a.NewCandidate(&content.FinalParams, &[]byte{})
	case SendTransactionEndpoint:
		return a.SendTransaction(&content.FinalParams, &[]byte{})
	case getContactsListEndpoint:
		return a.GetContactList(&content.FinalParams, &content.FinalReply)
	case TestServerEndpoint:
		return a.TestServer(&content.FinalParams, &content.FinalReply)
	case newMessageEndpoint:
		return a.NewCaesarMessage(&content.FinalParams, &[]byte{})
	}
	return nil
}
func (a *AdamniteServer) callOnSelfThenShare(content ForwardingContent) error {
	log.Println("call on self and share")
	if err := a.callOnSelf(content); err != nil {
		log.Printf(serverPreface, fmt.Sprintf("Error: %s", err))
		if err != ErrBadForward {
			//most of the time, an error for us, isn't an error for all. This is to stop a message from being forwarded at all.
			return nil
		}
		return err
	}
	return a.forwardingMessageReceived(content, &content.FinalReply)
}

const getContactsListEndpoint = "AdamniteServer.GetContactList"

func (a *AdamniteServer) GetContactList(params *[]byte, reply *[]byte) (err error) {
	a.print("Get Contact list")
	contacts := a.GetContactsFunction()
	*reply, err = encoding.Marshal(contacts)
	return
}

const getVersionEndpoint = "AdamniteServer.GetVersion"

func (a *AdamniteServer) GetVersion(params *[]byte, reply *[]byte) error {
	a.print("Get Version")
	receivedData := struct {
<<<<<<< Updated upstream
		Address           common.Address
		HostingServerPort string
=======
		Address                 bytes.Address
		HostingServerConnection string
>>>>>>> Stashed changes
	}{}
	if err := encoding.Unmarshal(*params, &receivedData); err != nil {
		a.printError("Get Version", err)
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
	ans.Addr_from = a.hostingNodeID
	// ans.Last_round = a.chain.CurrentBlock().Number()
	if data, err := encoding.Marshal(ans); err != nil {
		a.printError("Get Version", err)
		return err
	} else {
		*reply = data
	}
	return nil
}

const getChainIDEndpoint = "AdamniteServer.GetChainID"

func (a *AdamniteServer) GetChainID(params *[]byte, reply *[]byte) error {
	a.print("Get chain ID")
	if a.chain == nil || a.chain.Config() == nil {
		return ErrChainNotSet
	}

	data, err := encoding.Marshal(a.chain.Config().ChainID.String())
	if err != nil {
		a.printError("Get chain ID", err)
		return err
	}

	*reply = data
	return nil
}

const getBalanceEndpoint = "AdamniteServer.GetBalance"

func (a *AdamniteServer) GetBalance(params *[]byte, reply *[]byte) error {
	a.print("Get balance")
	input := struct {
		Address string
	}{}

	if err := encoding.Unmarshal(*params, &input); err != nil {
		a.printError("Get balance", err)
		return err
	}

	data, err := encoding.Marshal(a.stateDB.GetBalance(common.HexToAddress(input.Address)).String())
	if err != nil {
		a.printError("Get balance", err)
		return err
	}

	*reply = data
	return nil
}

const getAccountsEndpoint = "AdamniteServer.GetAccounts"

func (a *AdamniteServer) GetAccounts(params *[]byte, reply *[]byte) error {
	a.print("Get accounts")

	data, err := encoding.Marshal(a.addresses)
	if err != nil {
		a.printError("Get accounts", err)
		return err
	}

	*reply = data
	return nil
}

const getBlockByHashEndpoint = "AdamniteServer.GetBlockByHash"

func (a *AdamniteServer) GetBlockByHash(params *[]byte, reply *[]byte) error {
	a.print("Get block by hash")

	input := struct {
		BlockHash common.Hash
	}{}

	if err := encoding.Unmarshal(*params, &input); err != nil {
		a.printError("Get block by hash", err)
		return err
	}

	data, err := encoding.Marshal(*a.chain.GetBlockByHash(input.BlockHash))
	if err != nil {
		a.printError("Get block by hash", err)
		return err
	}

	*reply = data
	return nil

}

const getBlockByNumberEndpoint = "AdamniteServer.GetBlockByNumber"

func (a *AdamniteServer) GetBlockByNumber(params *[]byte, reply *[]byte) error {
	a.print("Get block by number")

	input := struct {
		BlockNumber big.Int
	}{}

	if err := encoding.Unmarshal(*params, &input); err != nil {
		a.printError("Get block by number", err)
		return err
	}

	data, err := encoding.Marshal(*a.chain.GetBlockByNumber(&input.BlockNumber))
	if err != nil {
		a.printError("Get block by number", err)
		return err
	}

	*reply = data
	return nil
}

const createAccountEndpoint = "AdamniteServer.CreateAccount"

func (a *AdamniteServer) CreateAccount(params *[]byte, reply *[]byte) error {
	a.print("Create account")

	input := struct {
		Address string
	}{}

	if err := encoding.Unmarshal(*params, &input); err != nil {
		a.printError("Create account", err)
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

	data, err := encoding.Marshal(true)
	if err != nil {
		a.printError("Create account", err)
		return err
	}

	*reply = data
	return nil
}

const SendTransactionEndpoint = "AdamniteServer.SendTransaction"

func (a *AdamniteServer) SendTransaction(params *[]byte, reply *[]byte) error {
	a.print("Send transaction")
	// if a.stateDB == nil {
	// 	return ErrStateNotSet
	// }
	var input *utils.Transaction

	if err := encoding.Unmarshal(*params, &input); err != nil {
		a.printError("Send transaction", err)
		return err
	}
	data, err := encoding.Marshal(true)
	if err != nil {
		a.printError("Send transaction", err)
		return err
	}
	*reply = data

	if a.newTransactionReceived != nil {
		return a.newTransactionReceived(input, params)
	} else {
		//TODO: this node cant forward this transaction at all.
	}

	return nil

}

func NewAdamniteServer(stateDB *statedb.StateDB, chain *blockchain.Blockchain, port uint32) *AdamniteServer {
	rpcServer := rpc.NewServer()

	adamnite := new(AdamniteServer)
	adamnite.stateDB = stateDB
	adamnite.chain = chain
	adamnite.timesTestHasBeenCalled = 0
	adamnite.seenConnections = make(map[common.Hash]common.Void)
	adamnite.DebugOutput = false

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
				// _ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				rpcServer.ServeConn(conn)
			}(conn)
		}
	}
	adamnite.listener = listener
	adamnite.Run = runFunc
	return adamnite
}
