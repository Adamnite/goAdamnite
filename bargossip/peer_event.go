package bargossip

import (
	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

// PeerEventType is the type of peer events emitted by a p2p.Server
type PeerEventType string

const (
	// PeerEventTypeAdd is the type of event emitted when a peer is added
	// to a p2p.Server
	PeerEventTypeAdd PeerEventType = "add"

	// PeerEventTypeDrop is the type of event emitted when a peer is
	// dropped from a p2p.Server
	PeerEventTypeDrop PeerEventType = "drop"

	// PeerEventTypeMsgSend is the type of event emitted when a
	// message is successfully sent to a peer
	PeerEventTypeMsgSend PeerEventType = "msgsend"

	// PeerEventTypeMsgRecv is the type of event emitted when a
	// message is received from a peer
	PeerEventTypeMsgRecv PeerEventType = "msgrecv"
)

// PeerEvent is an event emitted when peers are either added or dropped from
// a p2p.Server or when a message is sent or received on a peer connection
type PeerEvent struct {
	Type          PeerEventType  `json:"type"`
	Peer          admnode.NodeID `json:"peer"`
	Error         string         `json:"error,omitempty"`
	ProtocolID    uint           `json:"protocol,omitempty"`
	MsgCode       *uint64        `json:"msg_code,omitempty"`
	MsgSize       *uint32        `json:"msg_size,omitempty"`
	LocalAddress  string         `json:"local,omitempty"`
	RemoteAddress string         `json:"remote,omitempty"`
}
