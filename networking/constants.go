package networking

import (
	"fmt"
	"time"

	"github.com/adamnite/go-adamnite/common"
)

var (
	ErrPreexistingConnection   = fmt.Errorf("contact already has active connection")
	ErrOutboundCapacityReached = fmt.Errorf("currently at capacity for outbound connections")
	ErrContactBlacklisted      = fmt.Errorf("contact attempting to be added is already blacklisted")
	ErrContactIsSelf           = fmt.Errorf("the contact you're trying to add is the owner of this contact list")
	ErrDistrustedConnection    = fmt.Errorf("the contact attempting to connect to is untrustworthy")
	ErrNoNewConnectionsMade    = fmt.Errorf("no new connections were actually made after sprawl")
)

var blacklisted common.Void

const (
	//the time we ascribe to a missed connection. We ascribe 1.5x the connection time limit.
	MISSED_CONNECTION_TIME_PENALTY int64  = int64(time.Duration(15 * time.Second))
	DISTRUST_CUTOFF                uint64 = 1000 //cutoff point where a connection is blacklisted
	TRUSTFUL_BENEFIT               uint64 = 5    //the amount of trust given back per truthful connection.
)

type NetworkTopLayerType uint8

const (
	NetworkingOnly        NetworkTopLayerType = 1 << iota //00000001 networking only
	PrimaryTransactions                                   //00000010 representing chamber A, or main transactions
	SecondaryTransactions                                 //00000100 representing chamber B, or VM consensus
	CaesarHandling                                        //00001000 handling caesar forwarding
)

// call on type you are checking for, on the test value. EG, call  NetworkingOnly.IsIn(NetAndPrimary)
func (nt NetworkTopLayerType) IsIn(test uint8) bool {
	return uint8(nt)&test == uint8(nt)
}

// IsTypeIn checks to see if the network type nt is in test.
func (nt NetworkTopLayerType) IsTypeIn(test NetworkTopLayerType) bool {
	return uint8(nt)&uint8(test) == uint8(nt)
}
