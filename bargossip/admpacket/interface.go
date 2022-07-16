package admpacket

import "github.com/adamnite/go-adamnite/bargossip/admnode"

type ADMPacket interface {
	Name() string        // Name returns a string corresponding to the message type.
	MessageType() byte   // Type returns the message type.
	RequestID() []byte   // Returns the request ID.
	SetRequestID([]byte) // Sets the request ID.
}

type StaticHeader struct {
	ProtocolID [10]byte
	Version    uint16
	Nonce      Nonce
	AuthSize   uint16
	PacketType byte
}

type Header struct {
	IV [sizeOfIV]byte

	StaticHeader

	AuthData []byte
	srcID    admnode.NodeID
}

type messageAuthData struct {
	SrcID admnode.NodeID
}

type askHandshakeAuthData struct {
	RandomID  [16]byte
	DposRound uint64
}

type handshakeAuthData struct {
	SrcID         admnode.NodeID
	SignatureSize byte
	PubkeySize    byte
	NodeInfo      []byte

	signature []byte
	pubkey    []byte
}
