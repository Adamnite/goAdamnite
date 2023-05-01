package networking

import (
	"fmt"
	"sort"

	"github.com/adamnite/go-adamnite/rpc"
)

type Contact struct { //the contacts list from this point.
	connectionString string //ip and port for the specified endpoint.
	NodeID           int
	//any other data needed about an endpoint would be stored here.
}

// keeps track of all the contacts, sorts them, and manages the connections.
type ContactBook struct {
	//an array of connections for general use, stored by pointer. Followed by mappings to give o1 performance of sorting static characteristics
	connections          []*connectionStatus
	connectionsByContact map[*Contact]*connectionStatus

	blacklist    []*Contact
	blacklistSet map[*Contact]void //taking a mapping to an empty struct doesn't take up any extra memory, and gives us o1 to check if a contact is blacklisted.
}

func NewContactBook() ContactBook {
	return ContactBook{
		connections:          make([]*connectionStatus, 0),
		connectionsByContact: make(map[*Contact]*connectionStatus),
		blacklist:            make([]*Contact, 0),
		blacklistSet:         make(map[*Contact]void),
	}
}

func (cb *ContactBook) AddConnection(contact *Contact) error {
	newConn := newConnectionStatus(contact)
	if _, blacklisted := cb.blacklistSet[contact]; blacklisted {
		return ErrContactBlacklisted
	}
	if len(cb.connections) != 0 {
		caseFound := false
		//we check if this contact is already in the contact list. If it isn't, we add it. this comes with more edge cases than id care to admit.
		//also, go doesn't allow bitwise manipulation from a boolean, so i cant convert this to more efficient(and fun looking) case.
		for i := 0; i < len(cb.connections) && !caseFound; i++ {
			switch sameConnectionString, sameNodeID := (contact.connectionString == cb.connections[i].contact.connectionString), (contact.NodeID == cb.connections[i].contact.NodeID); {
			case sameConnectionString && sameNodeID:
				//the id, and connection string are the same, therefor this has already been added
				fmt.Println("identical contact added")
				return nil
			case (!sameConnectionString && !sameNodeID) && i == len(cb.connections):
				//we've gone through the list, and not found this yet, therefor this has to be new
				caseFound = true
			case !sameConnectionString && sameNodeID:
				// we already have this nodeID in use, but under a different connection string.
				//TODO: figure out the best way to solve this. This could be caused by a node's router restarting, and changing its dynamic IP
				return fmt.Errorf("the contact you tried to connect to is currently saved under a different IP. Saved: %v, New: %v", cb.connections[i], contact)
			case sameConnectionString && !sameNodeID:
				// this is a new nodeID being sent from a saved connection string.
				//TODO: figure out the best way to solve this. This could be caused by a node being restarted and generating a new NodeID
				return fmt.Errorf("the contact you tried to connect to is currently saved under a different NodeID. Saved: %v, New: %v", cb.connections[i], contact)
			}
		}
	}
	cb.connections = append(cb.connections, newConn)
	cb.connectionsByContact[contact] = newConn
	return nil
}

// mark demerit points against this contact. Returns if you distrust them.(true means they have been blacklisted)
func (cb *ContactBook) Distrust(contact *Contact, amount uint64) bool {
	if cb.connectionsByContact[contact] == nil {
		cb.AddConnection(contact)
	}
	if !cb.connectionsByContact[contact].distrust(amount) {
		cb.AddToBlacklist(contact)
		return true
	}
	return false
}
func (cb *ContactBook) AddToBlacklist(contact *Contact) {
	if cb.connectionsByContact[contact] != nil {
		cb.connectionsByContact[contact] = nil
		for i := 0; i < len(cb.connections); i++ {
			if cb.connections[i].contact == contact {
				cb.connections = append(cb.connections[0:i-1], cb.connections[i:]...)
				break
			}
		}
	}
	if _, exists := cb.blacklistSet[contact]; exists {
		//already stored.
		return
	}
	cb.blacklist = append(cb.blacklist, contact)
	cb.blacklistSet[contact] = blacklisted
}
func (cb *ContactBook) GetContactList() rpc.PassedContacts {
	passed := rpc.PassedContacts{
		NodeIDs:                    []int{},
		ConnectionStrings:          []string{},
		BlacklistIDs:               []int{},
		BlacklistConnectionStrings: []string{},
	}
	for _, x := range cb.connections {
		passed.NodeIDs = append(passed.NodeIDs, x.contact.NodeID)
		passed.ConnectionStrings = append(passed.ConnectionStrings, x.contact.connectionString)
	}
	for _, x := range cb.blacklist {
		passed.BlacklistIDs = append(passed.BlacklistIDs, x.NodeID)
		passed.BlacklistConnectionStrings = append(passed.BlacklistConnectionStrings, x.connectionString)
	}

	return passed
}
func (conns ContactBook) GetAverageConnectionResponseTime() int64 {
	var total uint64 = 0 //use uint64 to hav extra space to work with.
	for _, x := range conns.connections {
		total += uint64(x.getAverageResponseTime())
	}
	return int64(total / uint64(len(conns.connections)))
}

// drops this percentage of the slowest responses. percentage should be passed as a range from 0-1
func (conns *ContactBook) DropSlowestPercentage(percentage float32) {
	cutoffCount := len(conns.connections) - int(float32(len(conns.connections))*(percentage))
	if percentage > 1 || percentage <= 0 || cutoffCount <= 0 {
		//if its over 1(or 100%) or negative, an error happened. If its 0(or drop 0%), were done! or if the cutoff count is less than actually removing any.
		return
	}

	if percentage == 1 { //drop all, so don't bother calculating order.
		conns.connections = make([]*connectionStatus, 0)
		conns.connectionsByContact = make(map[*Contact]*connectionStatus)
		return
	}
	sort.Slice(conns.connections, func(i, j int) bool {
		return conns.connections[i].getAverageResponseTime() < conns.connections[j].getAverageResponseTime()
	})

	conns.connections = conns.connections[0:cutoffCount]
}

type connectionStatus struct {
	//I store the connection status and history of a Contact.
	contact            *Contact //the contact this connection is about
	responseTimes      []int64  //the time it takes for each contact to respond, in nanoseconds.
	connectionAttempts int64    //number of times we've tried to get the connection
	demerits           uint64   //the number of "points" against this connections. If a connection is consistently misbehaving, it will quickly be removed. assume this is as grading over 1000. We don't overly punish network hiccups.
}

func newConnectionStatus(contact *Contact) *connectionStatus {
	return &connectionStatus{
		contact:            contact,
		responseTimes:      []int64{},
		connectionAttempts: 0,
		demerits:           0,
	}
}

// returns the average response time it has taken for this connection. Returns -1 if no connectionAttempts hav been made.
func (con *connectionStatus) getAverageResponseTime() int64 {
	if con.connectionAttempts == 0 {
		return -1
	}
	var sum int64 = 0
	for _, x := range con.responseTimes {
		sum += x
	}
	sum += (con.connectionAttempts - int64(len(con.responseTimes))) * MISSED_CONNECTION_TIME_PENALTY

	return sum / con.connectionAttempts
}

// a successful connection has been made. This is used to prevent someone for getting blacklisted for wifi hiccups.
func (con *connectionStatus) trust() {
	if con.demerits >= TRUSTFUL_BENEFIT {
		con.demerits -= TRUSTFUL_BENEFIT
	}
}

// used for a network member misbehaving. damage is the amount of distrust they have caused, returns true if they are still usable.
func (con *connectionStatus) distrust(damage uint64) bool {
	con.demerits += damage
	return con.demerits >= DISTRUST_CUTOFF
}
