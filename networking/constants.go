package networking

import (
	"fmt"
	"time"
)

var (
	err_preexistingConnection   = fmt.Errorf("contact already has active connection")
	err_outboundCapacityReached = fmt.Errorf("currently at capacity for outbound connections")
	err_contact_blacklisted     = fmt.Errorf("contact attempting to be added is already blacklisted")
)

type void struct{}

var blacklisted void

const (
	//the time we ascribe to a missed connection. We ascribe 1.5x the connection time limit.
	MISSED_CONNECTION_TIME_PENALTY = int64(time.Duration(15 * time.Second))
	DISTRUST_CUTOFF                = uint64(1000) //cutoff point where a connection is blacklisted
	TRUSTFUL_BENEFIT               = uint64(5)    //the amount of trust given back per truthful connection.
)
