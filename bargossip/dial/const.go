package dial

import "time"

type ConnectionFlag int32

const (
	defaultMaxOutboundConnections        = 15
	defaultMaxPendingOutboundConnections = 15

	defaultDialHisotryExpiration = 30 * time.Second
	defaultDialTimeout           = 20 * time.Second
)

const (
	InboundConnection ConnectionFlag = 1 << iota
	OutboundConnection
)
