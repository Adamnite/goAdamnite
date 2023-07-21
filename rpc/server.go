package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strings"
	"time"

	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils"
	encoding "github.com/vmihailenco/msgpack/v5"
)

type AdamniteServer struct {
	chain           *blockchain.Blockchain
	hostingNodeID   utils.Address
	seenConnections map[utils.Hash]utils.Void
	Version         string

	GetContactsFunction       func() PassedContacts
	listener                  net.Listener
	mostRecentReceivedIP      string //TODO: CHECK THIS! Most likely can cause a race condition.
	timesTestHasBeenCalled    int
	newConnection             func(string, utils.Address)
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
	newConn func(string, utils.Address),
	newTransaction func(*utils.Transaction, *[]byte) error) {
	a.forwardingMessageReceived = newForward
	a.newConnection = newConn
	a.newTransactionReceived = newTransaction
}
func (a *AdamniteServer) SetForwardFunc(newForward func(ForwardingContent, *[]byte) error) {
	a.forwardingMessageReceived = newForward
}
func (a *AdamniteServer) SetNewConnectionFunc(newConn func(string, utils.Address)) {
	a.newConnection = newConn
}

// set a response point if we get asked to handle a transaction
func (a *AdamniteServer) SetTransactionHandler(handler func(*utils.Transaction, *[]byte) error) {
	a.newTransactionReceived = handler
}

func (a *AdamniteServer) SetHostingID(id *utils.Address) {
	if id == nil {
		a.hostingNodeID = utils.Address{0}
		return
	}
	a.hostingNodeID = *id
}
func (a *AdamniteServer) Close() {
	//TODO: clear all mappings!
	for h := range a.seenConnections {
		delete(a.seenConnections, h)
	}
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
	a.seenConnections[fc.Hash()] = utils.Void{} //this doesn't actually use any memory
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
		Address           utils.Address
		HostingServerPort string
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

func NewAdamniteServer(port uint32) *AdamniteServer {
	rpcServer := rpc.NewServer()

	adamnite := new(AdamniteServer)
	adamnite.timesTestHasBeenCalled = 0
	adamnite.seenConnections = make(map[utils.Hash]utils.Void)
	adamnite.DebugOutput = false
	adamnite.Version = "0.1.2"

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
