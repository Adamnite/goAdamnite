package networking

import (
	"log"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/rpc"
	"github.com/adamnite/go-adamnite/utils"
	"golang.org/x/sync/syncmap"
)

//version packet needs to contain
//client_version
//timestamp
//addr_received
//addr_from
//last_round
//nonce

// once a node gets a whitelist from a seed node(most likely but not required), it gets the block information from those nodes.
// these nodes sending the data are useful, as this can act as a test to see if they are still active (eg, if no response is heard from them after 10s, you can remove them from the whitelist)
// when a node receives a new list, it assumes that they are "gray" nodes, testing their response time with a handshake, if these nodes respond within time, they are added to the whitelist.
//
// Node needs to be able to update its gray list, which should be reflected in the RPC servers answers.
// to do this, we could use a local DB instance to cache all known nodes, then update the db here, but get the data in RPC?
// pass RPC an array so we can still change the contact list without actually needing to change anything, or keep reference to the RPC object?
//
//
//adding a blacklist will be required within this to ban bad acting members from clogging the server

// NetNode does not handle whitelisting or blacklisting, only the interactions with specified points.
type NetNode struct {
	thisContact Contact
	contactBook ContactBook //list of known contacts. Assume this to be gray.

	maxOutboundConnections uint        //how many outbound connections can this reply to.
	activeOutboundCount    uint        //how many connections are active
	activeContactToClient  syncmap.Map //spin up a new client for each outbound connection. *contact -> *rpc.AdamniteClient

	hostingServer *rpc.AdamniteServer
	bouncerServer *rpc.BouncerServer //bouncer server is optional. Acting as the input for off chain interactions (eg, from web)

	consensusCandidateHandler   func(utils.Candidate) error
	consensusVoteHandler        func(utils.Voter) error
	consensusTransactionHandler func(*utils.Transaction) error
	consensusBlockHandler       func(utils.Block) error
}

// TODO: we should add a port option so that people can allow forwarding on that port
func NewNetNode(address common.Address) *NetNode {
	n := NetNode{
		thisContact:            Contact{NodeID: address}, //TODO: add the address on netNode creation.
		maxOutboundConnections: 5,
		activeOutboundCount:    0,
		activeContactToClient:  syncmap.Map{},
	}
	n.contactBook = NewContactBook(&n.thisContact)

	return &n
}
func (n *NetNode) GetOwnContact() Contact {
	return n.thisContact
}
func (n NetNode) GetConnectionString() string {
	if n.hostingServer == nil {
		return ""
	}
	return n.thisContact.ConnectionString
}
func (n NetNode) GetBouncerString() string {
	if n.bouncerServer == nil {
		return ""
	}
	return n.bouncerServer.Addr()

}

func (n *NetNode) SetMaxConnections(newMax uint) {
	n.maxOutboundConnections = newMax
}
func (n *NetNode) SetBounceServerMessaging(getMsgs func(common.Address, common.Address) []*utils.CaesarMessage) {
	if n.bouncerServer == nil {
		return
	}
	n.bouncerServer.SetMessagingHandlers(getMsgs)
}
func (n *NetNode) AddBouncerServer(
	state *statedb.StateDB, chain *blockchain.Blockchain,
	hostPort uint32,
) {
	if n.bouncerServer != nil {
		log.Println("closing old bouncer server before starting new")
		n.bouncerServer.Close()
	}
	n.bouncerServer = rpc.NewBouncerServer(state, chain, hostPort)
	n.bouncerServer.SetHandlers(n.handleForward)
}

// spins up a RPC server with chain reference, and capability to properly propagate transactions
func (n *NetNode) AddFullServer(
	state *statedb.StateDB, chain *blockchain.Blockchain,
	transactionHandler func(*utils.Transaction) error,
	blockHandler func(utils.Block) error,
	candidateHandler func(utils.Candidate) error,
	voteHandler func(utils.Voter) error) error {
	if n.hostingServer != nil {
		log.Println("closing old server before adding new server")
		n.hostingServer.Close() //assume they want to restart the server then
	}
	n.hostingServer = rpc.NewAdamniteServer(0)
	n.consensusCandidateHandler = candidateHandler
	n.consensusVoteHandler = voteHandler
	n.updateServer()
	n.consensusTransactionHandler = transactionHandler
	n.consensusBlockHandler = blockHandler

	return nil
}

func (n *NetNode) AddMessagingCapabilities(msgHandler func(*utils.CaesarMessage)) {
	if n.hostingServer == nil {
		n.AddServer()
	}
	n.hostingServer.SetCaesarMessagingHandlers(msgHandler)
}

// spins up a server for this node.
func (n *NetNode) AddServer() error {
	if n.hostingServer != nil {
		log.Println("closing old server before adding new server")
		n.hostingServer.Close() //assume they want to restart the server then
	}
	n.hostingServer = rpc.NewAdamniteServer(0)
	n.updateServer()
	return nil
}

func (n *NetNode) updateServer() {
	n.hostingServer.GetContactsFunction = n.contactBook.GetContactList
	n.hostingServer.SetHostingID(&n.thisContact.NodeID)
	n.hostingServer.SetHandlers(
		n.handleForward,
		n.versionCheck,
		n.handleTransaction,
		n.handleBlock,
	)
	n.hostingServer.SetConsensusHandlers(
		n.consensusCandidateHandler,
		n.consensusVoteHandler,
	)
	n.hostingServer.Start()
	n.thisContact.ConnectionString = n.hostingServer.Addr()
}

func (n *NetNode) Close() {
	if n.hostingServer != nil {
		n.hostingServer.Close()
	}
	n.activeContactToClient.Range(func(key, value any) bool {
		active := value.(*rpc.AdamniteClient)
		active.Close()
		n.activeContactToClient.Delete(key)
		return true
	})
	n.contactBook.Close()

}

// use to setup a max length a node will have its grey list as. Use 0 to ignore this. Only truncates when shortening the list
func (n *NetNode) SetMaxGreyList(maxLength uint) {
	n.contactBook.maxGreyList = maxLength
}
func (n *NetNode) handleBlock(block utils.Block) error {
	//TODO: if you wanted to log all blocks, even invalid ones, you would do so here
	if n.consensusBlockHandler == nil {
		return nil
	}
	return n.consensusBlockHandler(block)
}
func (n *NetNode) handleTransaction(transaction *utils.Transaction) error {
	//TODO: here is where a logging method could be nice for anyone looking to track all transactions, including failed ones
	if n.consensusTransactionHandler == nil {
		//we can't verify this, so just propagate it out!
		return nil
	}
	return n.consensusTransactionHandler(transaction)
}
func (n *NetNode) handleForward(content rpc.ForwardingContent, reply *[]byte) error {
	n.hostingServer.AlreadySeen(content) //just make sure (especially if we're calling this ourselves) we don't get a recalling of this.
	// log.Println("handling forwarding")//TODO: useful for debugging. When we setup a proper debugger, please fix this
	if content.DestinationNode != nil {
		//see if we're actively connected to them, then send it as a direct call to them. (still use forward)
		var ansErr error
		n.activeContactToClient.Range(func(key, value any) bool {
			if key.(*Contact).NodeID == *content.DestinationNode {
				ansErr = value.(*rpc.AdamniteClient).ForwardMessage(content, reply)
				return false
			}
			return true
		})
		return ansErr
	}
	//this has been added to us, (and isn't called if the message is directly to us.)
	// log.Printf("forwarding to all %v known contacts", len(n.activeContactToClient)) //TODO: useful for debugging. When we setup a proper debugger, please fix this
	var errs error = nil
	n.activeContactToClient.Range(func(key, value any) bool {
		element := value.(*rpc.AdamniteClient)
		if err := element.ForwardMessage(content, &[]byte{}); err != nil {
			if err.Error() != rpc.ErrAlreadyForwarded.Error() {
				log.Println(err)
				//networking errors sometimes get weird, check the err.Error()
				//with a complex web, its likely one node already heard the message, no need to panic
				errs = err
				return false
			}
		}
		return true
	})
	return errs
}

// TODO: eventually this will also make sure versioning is the same, or will be renamed.
func (n *NetNode) versionCheck(remoteIP string, nodeID common.Address) {
	if remoteIP == "" {
		return
	}
	n.contactBook.AddConnection(&Contact{ConnectionString: remoteIP, NodeID: nodeID})
}

func (n *NetNode) GetConnectionsContacts(contact *Contact) error {
	connection, preexisting := n.activeContactToClient.Load(contact)
	if !preexisting {
		if err := n.ConnectToContact(contact); err != nil {
			return err
		}
		connection, _ = n.activeContactToClient.Load(contact)
	}
	//we can assume we always have a *working* connection from here onwards.
	contactListGiven := connection.(*rpc.AdamniteClient).GetContactList()
	if len(contactListGiven.NodeIDs) != len(contactListGiven.ConnectionStrings) || len(contactListGiven.BlacklistIDs) != len(contactListGiven.BlacklistConnectionStrings) {
		//the node has given an inaccurate node list/blacklist.
		if n.contactBook.Distrust(contact, 500) {
			return n.DropConnection(contact)
		}
		//TODO: this should return an error, but realistically you could just try them again. So should have an error to consider that.
		return nil
	}
	for i, id := range contactListGiven.NodeIDs {
		n.contactBook.AddConnection(&Contact{
			NodeID:           id,
			ConnectionString: contactListGiven.ConnectionStrings[i],
		})
	}
	for i, id := range contactListGiven.BlacklistIDs {
		n.contactBook.AddToBlacklist(&Contact{
			NodeID:           id,
			ConnectionString: contactListGiven.BlacklistConnectionStrings[i],
		})
	}
	return nil
}

// share a forward-able message across the network
func (n *NetNode) Propagate(v interface{}) error {
	foo, err := rpc.CreateForwardToAll(v)
	if err != nil {
		return err
	}
	return n.handleForward(foo, nil)
}
