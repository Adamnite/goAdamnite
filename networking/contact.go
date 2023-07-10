package networking

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/rpc"
)

type Contact struct { //the contacts list from this point.
	ConnectionString string //ip and port for the specified endpoint.
	NodeID           common.Address
	//any other data needed about an endpoint would be stored here.
}

// keeps track of all the contacts, sorts them, and manages the connections.
type ContactBook struct {
	ownerContact *Contact //useful so you don't try to add yourself to your own contacts list

	maxGreyList uint //does what it says on the tin. set to 0 to have unlimited size
	//an array of connections for general use, stored by pointer. Followed by mappings to give o1 performance of sorting static characteristics
	connections          []*connectionStatus
	connectionsByContact map[*Contact]*connectionStatus

	blacklistSet map[*Contact]common.Void //taking a mapping to an empty struct doesn't take up any extra memory, and gives us o1 to check if a contact is blacklisted.
}

func NewContactBook(owner *Contact) ContactBook {
	return ContactBook{
		ownerContact:         owner,
		maxGreyList:          0,
		connections:          make([]*connectionStatus, 0),
		connectionsByContact: make(map[*Contact]*connectionStatus),
		blacklistSet:         make(map[*Contact]common.Void),
	}
}
func (cb *ContactBook) Erase(contact *Contact) {
	if cb.ownerContact == contact {
		return //don't try to erase yourself.
	}
	cb.AddToBlacklist(contact)       // find that connection, then blacklist them so we can easily a single copy.
	delete(cb.blacklistSet, contact) //if they were blacklisted, now they aren't
}

// clears all records held here and deletes reference to itself.
func (cb *ContactBook) Close() {
	for c := range cb.blacklistSet {
		delete(cb.blacklistSet, c)
	}
	for c := range cb.connectionsByContact {
		delete(cb.connectionsByContact, c)
	}
	cb = nil
}

func (cb ContactBook) GetContactByEndpoint(endpoint string) *Contact {
	for c, _ := range cb.connectionsByContact {
		if c.ConnectionString == endpoint {
			return c
		}
	}
	return nil
}

func (cb *ContactBook) AddConnection(contact *Contact) error {
	if cb.ownerContact == contact {
		return ErrContactIsSelf //don't try to connect to yourself.
	}
	newConn := newConnectionStatus(contact)
	if _, blacklisted := cb.blacklistSet[contact]; blacklisted {
		return ErrContactBlacklisted
	}
	if len(cb.connections) != 0 {
		caseFound := false
		//we check if this contact is already in the contact list. If it isn't, we add it. this comes with more edge cases than id care to admit.
		//also, go doesn't allow bitwise manipulation from a boolean, so i cant convert this to more efficient(and fun looking) case.
		for i := 0; i < len(cb.connections) && !caseFound; i++ {
			switch sameConnectionString, sameNodeID := (contact.ConnectionString == cb.connections[i].contact.ConnectionString), (contact.NodeID == cb.connections[i].contact.NodeID); {
			case sameConnectionString && sameNodeID:
				//the id, and connection string are the same, therefor this has already been added
				// fmt.Println("identical contact added")
				*contact = *cb.connections[i].contact
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

func (cb *ContactBook) AddConnectionStatus(contact *Contact, timeDifference *time.Duration) {
	status, exists := cb.connectionsByContact[contact]
	if !exists {
		return
	}
	if timeDifference != nil {
		status.responseTimes = append(status.responseTimes, timeDifference.Nanoseconds())
		cb.Distrust(contact, 100)
	} else {
		status.trust()
	}
	status.connectionAttempts++
}

// mark demerit points against this contact. Returns if you distrust them.(true means they have been blacklisted)
func (cb *ContactBook) Distrust(contact *Contact, amount uint64) bool {
	if _, exists := cb.blacklistSet[contact]; exists {
		return true
	}
	if cb.connectionsByContact[contact] == nil {
		cb.AddConnection(contact)
	}
	if cs, exists := cb.connectionsByContact[contact]; exists && cs.distrust(amount) {
		cb.AddToBlacklist(contact)
		return true
	}
	return false
}
func (cb *ContactBook) AddToBlacklist(contact *Contact) {
	if cb.connectionsByContact[contact] != nil {
		delete(cb.connectionsByContact, contact)
		for i := 0; i < len(cb.connections); i++ {
			if cb.connections[i].contact == contact {
				if i == 0 {
					cb.connections = cb.connections[1:]
				} else {
					cb.connections = append(cb.connections[0:i-1], cb.connections[i:]...)
				}

				break
			}
		}
	}
	if _, exists := cb.blacklistSet[contact]; exists {
		//already stored.
		return
	}
	cb.blacklistSet[contact] = blacklisted
}
func (cb *ContactBook) GetContactList() rpc.PassedContacts {
	passed := rpc.PassedContacts{
		NodeIDs:                    []common.Address{},
		ConnectionStrings:          []string{},
		BlacklistIDs:               []common.Address{},
		BlacklistConnectionStrings: []string{},
	}
	for _, x := range cb.connections {
		passed.NodeIDs = append(passed.NodeIDs, x.contact.NodeID)
		passed.ConnectionStrings = append(passed.ConnectionStrings, x.contact.ConnectionString)
	}
	for blv, _ := range cb.blacklistSet {
		passed.BlacklistIDs = append(passed.BlacklistIDs, blv.NodeID)
		passed.BlacklistConnectionStrings = append(passed.BlacklistConnectionStrings, blv.ConnectionString)
	}

	return passed
}
func (cb ContactBook) GetAverageConnectionResponseTime() int64 {
	var total uint64 = 0 //use uint64 to hav extra space to work with.
	for _, x := range cb.connections {
		total += uint64(x.getAverageResponseTime())
	}
	return int64(total / uint64(len(cb.connections)))
}

// drops this percentage of the slowest responses. percentage should be passed as a range from 0-1
func (cb *ContactBook) DropSlowestPercentage(percentage float32) {
	cutoffCount := len(cb.connections) - int(float32(len(cb.connections))*(percentage))
	if percentage > 1 || percentage <= 0 || cutoffCount <= 0 {
		//if its over 1(or 100%) or negative, an error happened. If its 0(or drop 0%), were done! or if the cutoff count is less than actually removing any.
		return
	}

	if percentage == 1 { //drop all, so don't bother calculating order.
		cb.connections = make([]*connectionStatus, 0)
		cb.connectionsByContact = make(map[*Contact]*connectionStatus)
		return
	}
	cb.sortGreylist()

	cb.connections = cb.connections[0:cutoffCount]
	if cb.maxGreyList != 0 && len(cb.connections) > int(cb.maxGreyList) {
		cb.connections = cb.connections[0:cb.maxGreyList]
	}
}

func (cb *ContactBook) SelectWhitelist(goalCount int) []*Contact {
	cb.sortGreylist()
	connectionsLength := len(cb.connections)
	if goalCount*10 <= connectionsLength {
		//we only consider 10x the goal count connections.
		connectionsLength = goalCount * 10
	} else if goalCount >= connectionsLength {
		goalCount = connectionsLength
	}
	ansContacts := []*Contact{}
	attemptedIndexes := make(map[int]common.Void)
	for i := 0; len(ansContacts) < goalCount; i++ {

		index := int(math.Pow(rand.Float64(), 4) * float64(connectionsLength))
		if _, used := attemptedIndexes[index]; !used {
			//this is, in fact, a new index.
			ansContacts = append(ansContacts, cb.connections[index].contact)
			attemptedIndexes[index] = common.Void{}
		}
	}
	for k := range attemptedIndexes {
		delete(attemptedIndexes, k)
	}

	return ansContacts
}

// sorts the contacts by their average response time
func (cb *ContactBook) sortGreylist() {
	avgResponseTimes := map[int]int64{}
	sort.Slice(cb.connections, func(i, j int) bool {
		a, exists := avgResponseTimes[i]
		if !exists {
			a = cb.connections[i].getAverageResponseTime()
			avgResponseTimes[i] = a
		}
		b, exists := avgResponseTimes[j]
		if !exists {
			b = cb.connections[j].getAverageResponseTime()
			avgResponseTimes[j] = b
		}
		return a < b
	})
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
