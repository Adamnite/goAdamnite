package networking

import (
	"fmt"
	"math/rand"

	"github.com/adamnite/go-adamnite/rpc"
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

	maxInboundConnections  uint                             //how many inbound connections can this reply to.
	maxOutboundConnections uint                             //how many outbound connections can this reply to.
	activeOutboundCount    uint                             //how many connections are active
	activeContactToClient  map[*Contact]*rpc.AdamniteClient //spin up a new client for each outbound connection.

	hostingServer *rpc.AdamniteServer
}

func NewNetNode() *NetNode {
	n := NetNode{
		thisContact:            Contact{NodeID: rand.Int()},
		maxInboundConnections:  5,
		maxOutboundConnections: 5,
		activeOutboundCount:    0,
		activeContactToClient:  make(map[*Contact]*rpc.AdamniteClient),
	}
	n.contactBook = NewContactBook(&n.thisContact)

	return &n
}

// spins up a server for this node.
func (n *NetNode) AddServer() error {
	admServer := rpc.NewAdamniteServer(nil, nil, 0) //TODO: pass more parameters to this.

	admServer.GetContactsFunction = n.contactBook.GetContactList
	n.thisContact.connectionString = admServer.Addr()
	go admServer.Run()
	n.hostingServer = admServer

	return nil
}

func (n *NetNode) ConnectToContact(contact *Contact) error {
	if n.activeContactToClient[contact] != nil {
		return ErrPreexistingConnection
	}
	if n.activeOutboundCount >= n.maxOutboundConnections {
		return ErrOutboundCapacityReached
	}
	if err := n.contactBook.AddConnection(contact); err != nil {
		return err
	}

	if newClient, err := rpc.NewAdamniteClient(contact.connectionString); err != nil {
		return err
	} else {
		n.activeContactToClient[contact] = &newClient
		n.activeOutboundCount++
		return nil
	}
}

func (n *NetNode) DropConnection(contact *Contact) error {
	if n.activeContactToClient[contact] == nil {
		return fmt.Errorf("contact is not currently connected")
	}
	n.activeContactToClient[contact].Close()
	delete(n.activeContactToClient, contact)
	n.activeOutboundCount--
	return nil
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
			connectionString: contactListGiven.ConnectionStrings[i],
		})
	}
	for i, id := range contactListGiven.BlacklistIDs {
		n.contactBook.AddToBlacklist(&Contact{
			NodeID:           id,
			connectionString: contactListGiven.BlacklistConnectionStrings[i],
		})
	}
	return nil
}

// get the node to go around asking everyone it knows, who they know (layers times), and ignoring the autoCutoff slowest ones (as a percentage, range [0,1])
// note, this does temporarily remove the current connections. and may take a while to run
func (n *NetNode) SprawlConnections(layers int, autoCutoff float32) error {
	startingConnections := []*Contact{}
	if autoCutoff >= 1 || autoCutoff < 0 {
		autoCutoff = 0
	}
	for contact := range n.activeContactToClient {
		startingConnections = append(startingConnections, contact)
		n.DropConnection(contact)
	}
	talkedToContacts := make(map[*Contact]bool)
	talkedToContacts[&n.thisContact] = true
	//there are now no active connections.
	for i := 0; i < layers; i++ {
		//this is done EACH layer.
		for _, con := range n.contactBook.connections {
			if !talkedToContacts[con.contact] {
				// n.ConnectToContact(con.contact)
				n.GetConnectionsContacts(con.contact)
				n.DropConnection(con.contact)
				talkedToContacts[con.contact] = true
			}
		}
		n.contactBook.DropSlowestPercentage(autoCutoff / 2) //take half per layer, don't worry, the full removal happens at the end.
	}
	n.contactBook.DropSlowestPercentage(autoCutoff)
	for _, contact := range startingConnections {
		n.ConnectToContact(contact)
	}
	return nil
}
