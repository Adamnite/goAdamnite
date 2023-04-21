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
//
//
//
//adding a blacklist will be required within this to ban bad acting members from clogging the server

// NetNode does not handle whitelisting or blacklisting, only the interactions with specified points.
type NetNode struct {
	thisContact Contact
	contactList []*Contact //list of known contacts. Assume this to be gray.

	maxInboundConnections  uint                            //how many inbound connections can this reply to.
	maxOutboundConnections uint                            //how many outbound connections can this reply to.
	activeOutboundCount    uint                            //how many connections are active
	activeContactToClient  map[Contact]*rpc.AdamniteClient //spin up a new client for each outbound connection.

	hostingServers []*rpc.AdamniteServer
}

func NewNetNode() *NetNode {
	n := NetNode{
		thisContact:            Contact{NodeID: rand.Int()},
		contactList:            []*Contact{},
		maxInboundConnections:  5,
		maxOutboundConnections: 5,
		activeOutboundCount:    0,
		activeContactToClient:  make(map[Contact]*rpc.AdamniteClient),
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

func (n *NetNode) ConnectToContact(contact Contact) error {
	if n.activeContactToClient[contact] != nil {
		return err_preexistingConnection
	}
	if n.activeOutboundCount >= n.maxOutboundConnections {
		return err_outboundCapacityReached
	}
	if len(n.contactList) == 0 {
		n.contactList = append(n.contactList, &contact)
	} else {
		caseFound := false
		//we check if this contact is already in the contact list. If it isn't, we add it. this comes with more edge cases than id care to admit.
		//also, go doesn't allow bitwise manipulation from a boolean, so i cant convert this to more efficient(and fun looking) case.
		for i := 0; i < len(n.contactList) && !caseFound; i++ {
			switch sameConnectionString, sameNodeID := (contact.connectionString == n.contactList[i].connectionString), (contact.NodeID == n.contactList[i].NodeID); {
			case sameConnectionString && sameNodeID:
				//the id, and connection string are the same, therefor this has already been added
				fmt.Println("identical contact added")
			case (!sameConnectionString && !sameNodeID) && i == len(n.contactList):
				//we've gone through the list, and not found this yet, therefor this has to be new
				n.contactList = append(n.contactList, &contact)
				caseFound = true
			case !sameConnectionString && sameNodeID:
				// we already have this nodeID in use, but under a different connection string.
				//TODO: figure out the best way to solve this. This could be caused by a node's router restarting, and changing its dynamic IP
				return fmt.Errorf("the contact you tried to connect to is currently saved under a different IP. Saved: %v, New: %v", n.contactList[i], contact)
			case sameConnectionString && !sameNodeID:
				// this is a new nodeID being sent from a saved connection string.
				//TODO: figure out the best way to solve this. This could be caused by a node being restarted and generating a new NodeID
				return fmt.Errorf("the contact you tried to connect to is currently saved under a different NodeID. Saved: %v, New: %v", n.contactList[i], contact)
			}
		}
	}

	n.activeContactToClient[contact] = rpc.NewAdamniteClient(contact.connectionString)
	n.activeOutboundCount++
	return nil
}

func (n *NetNode) DropConnection(contact Contact) error {
	if n.activeContactToClient[contact] == nil {
		return fmt.Errorf("contact is not currently connected")
	}
	n.activeContactToClient[contact].Close()
	n.activeContactToClient[contact] = nil
	n.activeOutboundCount--
	return nil
}
