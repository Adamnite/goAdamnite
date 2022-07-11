package admpacket

import "github.com/adamnite/go-adamnite/gossip/admnode"

type ADMPacket interface {
	Name() string        // Name returns a string corresponding to the message type.
	MessageType() byte   // Type returns the message type.
	RequestID() []byte   // Returns the request ID.
	SetRequestID([]byte) // Sets the request ID.
}

type Findnode struct {
	RequestID []byte
	Distances []uint
}

type RspNodes struct {
	RequestID []byte
	Total     uint8
	Nodes     []*admnode.NodeInfo
}
