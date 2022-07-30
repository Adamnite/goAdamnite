package bargossip

import (
	"crypto/ecdsa"
	"net"
	"time"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/bargossip/dial"
)

const (
	handshakeTimeout            = 6 * time.Second
	AdamniteTCPHandshakeVersion = 1

	messageWriteTimeout = 20 * time.Second
	messageReadTimeout  = 20 * time.Second

	messagePayloadMaxSize = 1024 * 1024 * 16 // 16MB

	pingInterval = 5 * time.Second
)

// wrapPeerConnection is the wrapper to connection with the remote peer
type wrapPeerConnection struct {
	conn      net.Conn
	node      *admnode.GossipNode
	connFlags dial.ConnectionFlag
	protocol  *exchangeProtocol

	peerTransport

	// channels
	chError chan error
}

type MsgReader interface {
	ReadMsg() (Msg, error)
}

type MsgWriter interface {
	WriteMsg(Msg) error
}

type MsgReadWriter interface {
	MsgReader
	MsgWriter
}

// peerTransport is the interface to establish the transport between peers
type peerTransport interface {
	doHandshake(prvKey *ecdsa.PrivateKey) (*ecdsa.PublicKey, error)            // First step to exchange the key
	doExchangeProtocol(exchProto *exchangeProtocol) (*exchangeProtocol, error) // Second step to exchange the protocol

	MsgReader
	MsgWriter

	close(err error)
}
