package admpacket

import (
	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/utils/mclock"
)

// AskHandshake contains the handshake require information
type AskHandshake struct {
	HandshakeData []byte
	Nonce         Nonce
	RandomID      [16]byte
	DposRound     uint64
	Node          *admnode.GossipNode

	sent mclock.AbsTime
}

func (p *AskHandshake) Name() string {
	return "ADM-Handshake-Ask"
}

func (p *AskHandshake) MessageType() byte {
	return AskHandshakeMsg
}

func (p *AskHandshake) RequestID() []byte {
	return nil
}

func (p *AskHandshake) SetRequestID(id []byte) {

}
