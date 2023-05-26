package networking

import (
	"fmt"
	"strings"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/rpc"
)

// a file for the node's connection methods.

// reset a netNode to have the best connections available
func (n *NetNode) ResetConnections() error {
	if err := n.DropAllConnections(); err != nil {
		return err
	}
	return n.FillOpenConnections()
}

// make sure this node is running as many active connections as it can be
func (n *NetNode) FillOpenConnections() error {
	useRecursion := false
	if len(n.contactBook.connections) >= int(n.maxOutboundConnections*3) {
		if err := n.SprawlConnections(5, 0.01); err != nil {
			if err == errNoNewConnectionsMade {
				useRecursion = true
			} else {
				return err
			}
		}
	}
	possibleCons := n.contactBook.SelectWhitelist(int(n.maxOutboundConnections-n.activeOutboundCount) + 1)
	//get an extra incase one doesn't want to connect
	for i := 0; i < len(possibleCons) && n.activeOutboundCount < n.maxOutboundConnections; i++ {
		if err := n.ConnectToContact(possibleCons[i]); err != nil {
			switch err { //handle any handle-able errors directly here.
			case ErrPreexistingConnection:
				//since we have a recursion edge case, it is possible to attempt to call the same contact twice
			case ErrDistrustedConnection:
				// we tried connecting to them, and they stopped being truthful, well skip over them.
			case ErrOutboundCapacityReached:
				//someone probably ran this in an asynchronous thread, but our jobs done!
				return nil
			}
		}
	}
	if n.activeOutboundCount < n.maxOutboundConnections && useRecursion {
		return n.FillOpenConnections()
	}
	return nil
}
func (n *NetNode) ConnectToSeed(connectionPoint string) error {
	tempSeedContact := Contact{
		ConnectionString: connectionPoint,
		NodeID:           common.Address{0, 0, 0, 0}, //use the 0 address to test this
	}
	if err := n.ConnectToContact(&tempSeedContact); err != nil {
		return err
	}

	if err := n.GetConnectionsContacts(&tempSeedContact); err != nil { //get the seed nodes info so we can connect to them again more easily
		return err
	}
	n.contactBook.Erase(&tempSeedContact) //we don't have the full seed node information, so it's best to drop that.

	// do a full sprawl. This gets the seed node back and, given this is normally a part of the startup
	// lets us take our time and really fill our network
	if err := n.SprawlConnections(5, 0.05); err != nil {
		return err
	}
	return nil
}

// create an active connection to a known contact
func (n *NetNode) ConnectToContact(contact *Contact) error {
	if n.activeContactToClient[contact] != nil {
		return ErrPreexistingConnection
	} else if n.activeOutboundCount >= n.maxOutboundConnections {
		return ErrOutboundCapacityReached
	} else if err := n.contactBook.AddConnection(contact); err != nil {
		return err
	} else if n.contactBook.Distrust(contact, 0) {
		return ErrContactBlacklisted
	}

	if newClient, err := rpc.NewAdamniteClient(contact.ConnectionString); err != nil {
		return err
	} else {
		newClient.SetAddressAndHostingPort(
			&n.thisContact.NodeID,
			strings.Split(n.hostingServer.Addr(), ":")[1],
		)
		n.activeContactToClient[contact] = &newClient
		n.activeOutboundCount++
		working, err := n.testConnection(contact)
		if !working {
			n.DropConnection(contact)
			return ErrDistrustedConnection
		}
		return err
	}
}

// return wether a connection is worth using.
func (n *NetNode) testConnection(contact *Contact) (bool, error) {
	if n.activeContactToClient[contact] == nil {
		if err := n.ConnectToContact(contact); err != nil {
			return false, err
		}
	}
	connection := n.activeContactToClient[contact]
	timeBeforeConnect := time.Now().UTC()
	versionData, err := connection.GetVersion()
	timeAfterResponse := time.Now().UTC()

	if err != nil {
		n.contactBook.AddConnectionStatus(contact, nil)
		return false, err
	}
	// its /2 because this is recording the round trip response time.
	foo := timeAfterResponse.Sub(timeBeforeConnect) / 2
	n.contactBook.AddConnectionStatus(contact, &foo)
	//TODO: check the version running.
	if versionData.Timestamp.Before(timeBeforeConnect) || versionData.Timestamp.After(timeAfterResponse) {
		//trying to lie about when they received this.
		n.contactBook.Distrust(contact, 900)
	}
	if versionData.Addr_from != contact.NodeID {
		//returning the wrong nodeID, we almost entirely ban them, but we don't fully because this could have been misrepresented by someone else.
		n.contactBook.Distrust(contact, 900)
	}

	return true, nil
}

// drop all active outward connection
func (n *NetNode) DropAllConnections() error {
	for key := range n.activeContactToClient {
		if err := n.DropConnection(key); err != nil {
			return err
		}
	}
	return nil
}

// drop the active connection you have to this contact
func (n *NetNode) DropConnection(contact *Contact) error {
	if n.activeContactToClient[contact] == nil {
		return fmt.Errorf("contact is not currently connected")
	}
	n.activeContactToClient[contact].Close()
	delete(n.activeContactToClient, contact)
	n.activeOutboundCount--
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
		n.contactBook.DropSlowestPercentage(autoCutoff / (float32(layers) + 1)) //take half per layer, don't worry, the full removal happens at the end.
	}
	n.contactBook.DropSlowestPercentage(autoCutoff)
	for _, contact := range startingConnections {
		n.ConnectToContact(contact)
	}
	if len(talkedToContacts) == len(n.contactBook.connectionsByContact) {
		return errNoNewConnectionsMade
	}
	return nil
}
