package rpc

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"

	log "github.com/sirupsen/logrus"
	encoding "github.com/vmihailenco/msgpack/v5"
	"golang.org/x/sync/syncmap"
)

var USE_LOCAL_IP bool = false //set to true if you are testing locally and dont want to deal with 200 nodes trying to get the same IP
type AdamniteServer struct {
	chain           *blockchain.Blockchain
	hostingNodeID   common.Address
	externalIP      string
	seenConnections syncmap.Map // map[common.Hash]common.Void

	GetContactsFunction       func() PassedContacts
	listener                  net.Listener
	timesTestHasBeenCalled    int
	newConnection             func(string, common.Address)
	forwardingMessageReceived func(ForwardingContent, *[]byte) error
	newTransactionReceived    func(utils.TransactionType) error
	newBlockReceived          func(utils.BlockType) error
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
	newBlock func(utils.BlockType) error) {
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
	log.Debugf(serverPreface, methodName)
}
func (a *AdamniteServer) printError(methodName string, err error) {
	log.Errorf(serverPreface, fmt.Sprintf("%v\tError: %s", methodName, err))
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
		// if err != ErrBadForward {
		// 	//most of the time, an error for us, isn't an error for all. This is to stop a message from being forwarded at all.
		// 	return nil
		// }
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

const NewBlockEndpoint = "AdamniteServer.NewBlock"

func (a *AdamniteServer) NewBlock(params, reply *[]byte) error {
	a.print("New Block")
	if a.newBlockReceived == nil {
		//we aren't involved in blocks, so just pass it on
		data, err := encoding.Marshal(true)
		*reply = data
		return err
	}
	var basicBlock *utils.Block
	if err := encoding.Unmarshal(*params, &basicBlock); err != nil {
		return err
	}
	switch basicBlock.GetHeader().TransactionType {
	case utils.Transaction_Basic:
		return a.newBlockReceived(basicBlock)
	case utils.Transaction_VM_Call, utils.Transaction_VM_NewContract:
		var vmBlock *utils.VMBlock
		if err := encoding.Unmarshal(*params, &vmBlock); err != nil {
			return err
		} else {
			return a.newBlockReceived(vmBlock)
		}
	}

	return nil
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

func NewAdamniteServer(port uint32) *AdamniteServer {
	rpcServer := rpc.NewServer()

	adamnite := new(AdamniteServer)
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
