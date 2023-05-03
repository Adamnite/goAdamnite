package networking

import (
	"fmt"
	"time"
)

var (
	ErrPreexistingConnection   = fmt.Errorf("contact already has active connection")
	ErrOutboundCapacityReached = fmt.Errorf("currently at capacity for outbound connections")
	ErrContactBlacklisted      = fmt.Errorf("contact attempting to be added is already blacklisted")
	ErrContactIsSelf           = fmt.Errorf("the contact you're trying to add is the owner of this contact list")
	ErrDistrustedConnection    = fmt.Errorf("the contact attempting to connect to is untrustworthy")
)

type void struct{}

var blacklisted void

const (
	//the time we ascribe to a missed connection. We ascribe 1.5x the connection time limit.
	MISSED_CONNECTION_TIME_PENALTY int64  = int64(time.Duration(15 * time.Second))
	DISTRUST_CUTOFF                uint64 = 1000 //cutoff point where a connection is blacklisted
	TRUSTFUL_BENEFIT               uint64 = 5    //the amount of trust given back per truthful connection.
)
