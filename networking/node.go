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

	hostingServers []*rpc.AdamniteServer
}

func NewNetNode() *NetNode {
	n := NetNode{
		thisContact:            Contact{NodeID: rand.Int()},
		contactBook:            NewContactBook(),
		maxInboundConnections:  5,
		maxOutboundConnections: 5,
		activeOutboundCount:    0,
		activeContactToClient:  make(map[*Contact]*rpc.AdamniteClient),
		hostingServers:         []*rpc.AdamniteServer{},
	}

	return &n
}

func (n *NetNode) AddServer() error {
	l, rf := rpc.NewAdamniteServer(nil, nil) //TODO: pass more parameters to this.
	n.thisContact.connectionString = l.Addr().String()
	go rf() //TODO: replace all of this to return a server, and keep that in the hosting servers array.

	return nil
}

func (n *NetNode) ConnectToContact(contact *Contact) error {
	if n.activeContactToClient[contact] != nil {
		return err_preexistingConnection
	}
	if n.activeOutboundCount >= n.maxOutboundConnections {
		return err_outboundCapacityReached
	}
	if err := n.contactBook.AddConnection(contact); err != nil {
		return err
	}
	//TODO, use the contact book to check if it has this contact already.
	n.activeContactToClient[contact] = rpc.NewAdamniteClient(contact.connectionString)
	n.activeOutboundCount++
	return nil
}

func (n *NetNode) DropConnection(contact *Contact) error {
	if n.activeContactToClient[contact] == nil {
		return fmt.Errorf("contact is not currently connected")
	}
	n.activeContactToClient[contact].Close()
	n.activeContactToClient[contact] = nil
	n.activeOutboundCount--
	return nil
}
