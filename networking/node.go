package networking

import (
	"log"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/rpc"
	"github.com/adamnite/go-adamnite/utils"
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

	MaxOutboundConnections uint                             //how many outbound connections can this reply to.
	activeOutboundCount    uint                             //how many connections are active
	activeContactToClient  map[*Contact]*rpc.AdamniteClient //spin up a new client for each outbound connection.

	hostingServer *rpc.AdamniteServer

	consensusCandidateHandler   func(utils.Candidate) error
	consensusVoteHandler        func(utils.Voter) error
	consensusTransactionHandler func(*utils.Transaction) error
}

func NewNetNode(address common.Address) *NetNode {
	n := NetNode{
		thisContact:            Contact{NodeID: address}, //TODO: add the address on netNode creation.
		MaxOutboundConnections: 5,
		activeOutboundCount:    0,
		activeContactToClient:  make(map[*Contact]*rpc.AdamniteClient),
	}
	n.contactBook = NewContactBook(&n.thisContact)

	return &n
}
func (n NetNode) GetOwnContact() Contact {
	return n.thisContact
}

func (n NetNode) GetActiveConnectionsCount() uint {
	return n.activeOutboundCount
}
func (n NetNode) GetGreylistSize() int {
	return len(n.contactBook.connections)
}
func (n NetNode) GetMaxGreylist() int {
	return int(n.contactBook.maxGreyList)
}
func (n *NetNode) SetMaxConnections(newMax uint) {
	n.MaxOutboundConnections = newMax
}

// use to setup a max length a node will have its grey list as. Use 0 to ignore this. Only truncates when shortening the list
func (n *NetNode) SetMaxGreyList(maxLength uint) {
	n.contactBook.maxGreyList = maxLength
}

// spins up a RPC server with chain reference, and capability to properly propagate transactions
func (n *NetNode) AddFullServer(
	state *statedb.StateDB, chain *blockchain.Blockchain,
	transactionHandler func(*utils.Transaction) error,
	candidateHandler func(utils.Candidate) error,
	voteHandler func(utils.Voter) error) error {
	if n.hostingServer != nil {
		log.Println("closing old server before adding new server")
		n.hostingServer.Close() //assume they want to restart the server then
	}
	n.hostingServer = rpc.NewAdamniteServer(state, chain, 0)
	n.consensusCandidateHandler = candidateHandler
	n.consensusVoteHandler = voteHandler
	n.updateServer()
	n.consensusTransactionHandler = transactionHandler

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
	n.hostingServer = rpc.NewAdamniteServer(nil, nil, 0)
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
	)
	n.hostingServer.SetConsensusHandlers(
		n.consensusCandidateHandler,
		n.consensusVoteHandler,
	)
	go n.hostingServer.Run()
	n.thisContact.ConnectionString = n.hostingServer.Addr()
}

// closes the network, deletes all mappings, and deletes the NetNode
func (n *NetNode) Close() {
	if n.hostingServer != nil {
		n.hostingServer.Close()
	}
	for c := range n.activeContactToClient {
		n.activeContactToClient[c].Close()
		delete(n.activeContactToClient, c)
	}
	n.contactBook.Close()
	n = nil
}

func (n *NetNode) handleTransaction(transaction *utils.Transaction, transactionBytes *[]byte) error {
	if n.consensusTransactionHandler == nil {
		//we can't verify this, so just propagate it out!

		return n.handleForward(
			rpc.ForwardingContent{ //i don't love handling the forwarding generation here, but I'll live.
				FinalEndpoint: rpc.SendTransactionEndpoint,
				FinalParams:   *transactionBytes,
				InitialSender: transaction.From,
				// Signature:     common.BytesToHash(transaction.Signature), // i think this works, but not 100% sure its right.
			},
			&[]byte{},
		)
	}
	return n.consensusTransactionHandler(transaction)
}
func (n *NetNode) handleForward(content rpc.ForwardingContent, reply *[]byte) error {
	n.hostingServer.AlreadySeen(content) //just make sure (especially if we're calling this ourselves) we don't get a recalling of this.
	log.Println("handling forwarding")
	if content.DestinationNode != nil {
		//see if we're actively connected to them, then send it as a direct call to them. (still use forward)
		for key, connection := range n.activeContactToClient {
			if key.NodeID == *content.DestinationNode {
				return connection.ForwardMessage(content, reply)
			}
		}
	}
	//this has been added to us, (and isn't called if the message is directly to us.)
	log.Printf("forwarding to all %v known contacts", len(n.activeContactToClient))
	for _, element := range n.activeContactToClient {
		if err := element.ForwardMessage(content, &[]byte{}); err != nil {
			if err.Error() != rpc.ErrAlreadyForwarded.Error() {
				log.Println(err)
				//networking errors sometimes get weird, check the err.Error()
				//with a complex web, its likely one node already heard the message, no need to panic
				return err
			}
		}
	}
	return nil
}

// TODO: eventually this will also make sure versioning is the same, or will be renamed.
func (n *NetNode) versionCheck(remoteIP string, nodeID common.Address) {
	if remoteIP == "" {
		return
	}
	n.contactBook.AddConnection(&Contact{ConnectionString: remoteIP, NodeID: nodeID})
}

func (n *NetNode) GetConnectionsContacts(contact *Contact) error {
	if n.activeContactToClient[contact] == nil {
		if err := n.ConnectToContact(contact); err != nil {
			return err
		}
	}
	//we can assume we always have a *working* connection from here onwards.
	connection := n.activeContactToClient[contact]
	contactListGiven := connection.GetContactList()
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
