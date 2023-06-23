package rpc

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"
	encoding "github.com/vmihailenco/msgpack/v5"
	"golang.org/x/sync/syncmap"
)

var USE_LOCAL_IP bool = false //set to true if you are testing locally and dont want to deal with 200 nodes trying to get the same IP
type AdamniteServer struct {
	stateDB         *statedb.StateDB
	chain           *blockchain.Blockchain
	hostingNodeID   common.Address
	externalIP      string
	seenConnections syncmap.Map // map[common.Hash]common.Void

	addresses                 []string
	GetContactsFunction       func() PassedContacts
	listener                  net.Listener
	timesTestHasBeenCalled    int
	newConnection             func(string, common.Address)
	forwardingMessageReceived func(ForwardingContent, *[]byte) error
	newTransactionReceived    func(utils.TransactionType) error
	newBlockReceived          func(*utils.Block) error
	newCandidateHandler       func(utils.Candidate) error
	newVoteHandler            func(utils.Voter) error
	newMessageHandler         func(*utils.CaesarMessage)
	run                       func()
	currentlyRunning          bool
	DebugOutput               bool
}

func (a *AdamniteServer) Addr() string {
	return a.externalIP
}
func (a *AdamniteServer) SetHandlers(
	newForward func(ForwardingContent, *[]byte) error,
	newConn func(string, common.Address),
	newTransaction func(utils.TransactionType) error,
	newBlock func(*utils.Block) error) {
	a.forwardingMessageReceived = newForward
	a.newConnection = newConn
	a.newTransactionReceived = newTransaction
	a.newBlockReceived = newBlock
}
func (a *AdamniteServer) SetForwardFunc(newForward func(ForwardingContent, *[]byte) error) {
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
	a.currentlyRunning = false
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
	_, exists := a.seenConnections.LoadOrStore(fc.Hash(), true)
	return exists
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
	a.print("call on self")
	switch content.FinalEndpoint {
	case NewCandidateEndpoint:
		return a.NewCandidate(&content.FinalParams, &[]byte{})
	case NewVoteEndpoint:
		return a.NewVote(&content.FinalParams, &[]byte{})
	case NewTransactionEndpoint:
		return a.NewTransaction(&content.FinalParams, &[]byte{})
	case NewBlockEndpoint:
		return a.NewBlock(&content.FinalParams, &[]byte{})
	case getContactsListEndpoint:
		return a.GetContactList(&content.FinalParams, &content.FinalReply)
	case TestServerEndpoint:
		return a.TestServer(&content.FinalParams, &content.FinalReply)
	case newMessageEndpoint:
		return a.NewCaesarMessage(&content.FinalParams, &[]byte{})
	default:
		a.print("call on self, but no endpoint was found")
	}
	return nil
}
func (a *AdamniteServer) callOnSelfThenShare(content ForwardingContent) error {
	a.print("call on self and share")
	if err := a.callOnSelf(content); err != nil {
		a.printError("call on self and share", err)
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
		Address                 common.Address
		HostingServerConnection string
	}{}
	if err := encoding.Unmarshal(*params, &receivedData); err != nil {
		a.printError("Get Version", err)
		return err
	}
	if a.newConnection != nil && receivedData.HostingServerConnection != "" {
		//format it correctly to use the right port
		a.newConnection(receivedData.HostingServerConnection, receivedData.Address)
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

const NewBlockEndpoint = "AdamniteServer.NewBlock"

func (a *AdamniteServer) NewBlock(params, reply *[]byte) error {
	a.print("New Block")
	if a.newBlockReceived == nil {
		//we aren't involved in blocks, so just pass it on
		data, err := encoding.Marshal(true)
		*reply = data
		return err
	}
	var block *utils.Block
	if err := encoding.Unmarshal(*params, &block); err != nil {
		return err
	}
	return a.newBlockReceived(block)
}

const NewTransactionEndpoint = "AdamniteServer.NewTransaction"

func (a *AdamniteServer) NewTransaction(params *[]byte, reply *[]byte) error {
	a.print("New Transaction")
	var input *utils.BaseTransaction

	if err := encoding.Unmarshal(*params, &input); err != nil {
		a.printError("New Transaction", err)
		return err
	}
	//get the base transaction, just so we can check the type
	var transaction utils.TransactionType
	switch input.GetType() {
	case utils.Transaction_VM_Call:
		var vmTransaction *utils.VMCallTransaction
		if err := encoding.Unmarshal(*params, &vmTransaction); err != nil {
			a.printError("New Transaction", err)
			return err
		}
		transaction = vmTransaction
	case utils.Transaction_Basic:
		transaction = input
	}
	data, err := encoding.Marshal(true)
	if err != nil {
		a.printError("New Transaction", err)
		return err
	}
	*reply = data

	if a.newTransactionReceived != nil {
		return a.newTransactionReceived(transaction)
	}

	return nil
}

func NewAdamniteServer(stateDB *statedb.StateDB, chain *blockchain.Blockchain, port uint32) *AdamniteServer {
	rpcServer := rpc.NewServer()

	adamnite := new(AdamniteServer)
	adamnite.stateDB = stateDB
	adamnite.chain = chain
	adamnite.timesTestHasBeenCalled = 0
	adamnite.seenConnections = syncmap.Map{}
	adamnite.DebugOutput = false

	if err := rpcServer.Register(adamnite); err != nil {
		log.Fatal(err)
	}

	listener, _ := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if USE_LOCAL_IP {
		adamnite.externalIP = listener.Addr().String()
	} else {
		adamnite.externalIP = fmt.Sprintf("%s:%s", getExternalIP(), strings.Split(listener.Addr().String(), ":")[1])
	}
	log.Printf(serverPreface, fmt.Sprint("Endpoint: ", adamnite.externalIP))

	runFunc := func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("[Listener accept]", err)
				return
			}

			go func(conn net.Conn) {
				conn.RemoteAddr()
				defer func() {
					if err = conn.Close(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
						log.Println(err)
					}
				}()
				// adamnite.mostRecentReceivedIP = conn.RemoteAddr().String()
				//TODO: it is useful to get the IP of whoevers calling us, but god hates us, so nope...

				// _ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				rpcServer.ServeConn(conn)
			}(conn)
		}
	}
	adamnite.listener = listener
	adamnite.run = runFunc
	adamnite.currentlyRunning = false
	return adamnite
}
func (a *AdamniteServer) Start() {
	if a.currentlyRunning {
		//we don't want to just keep having the server run again and again
		return
	}
	a.currentlyRunning = true
	go a.run()
}

func getExternalIP() string {
	// url := "https://api.ipify.org" // we are using a pulib IP API, we're using ipify here, below are some others
	url := "http://myexternalip.com/raw"
	// http://myexternalip.com
	// http://api.ident.me
	// http://whatismyipaddress.com/api
	resp, err := http.Get(url)
	if err != nil {
		log.Println("error getting our IP")
		log.Fatal(err)
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println("error reading our IP")
		log.Fatal(err)
	}
	return string(ip)
}
