package adampro

import (
	"fmt"

	"github.com/adamnite/go-adamnite/bargossip"
	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

type Handler func(peer *Peer) error
type msgHandler func(admHandler AdamniteHandlerInterface, msg Decoder, peer *Peer) error

var admProtocolDemo = map[uint64]msgHandler{
	NewBlockMsg: newBlockHandler,
}

func MakeProtocols(adamniteHandler AdamniteHandlerInterface, chainID uint64, dnsdisc admnode.NodeIterator) []bargossip.SubProtocol {
	protocols := make([]bargossip.SubProtocol, len(ProtocolVersions))

	for i, v := range ProtocolVersions {
		protocols[i] = bargossip.SubProtocol{
			ProtocolID:         uint(i),
			ProtocolCodeOffset: v,
			ProtocolLength:     uint(ProtocolLength[v]),
			Run: func(p *bargossip.Peer, rw bargossip.MsgReadWriter) error {
				peer := NewPeer(v, p, rw, adamniteHandler.TxPool())
				defer peer.Close()
				return adamniteHandler.RunPeer(peer, func(peer *Peer) error {
					return ProtocolMsgHandler(adamniteHandler, peer)
				})
			},
		}
	}

	return protocols
}

func ProtocolMsgHandler(admHandler AdamniteHandlerInterface, peer *Peer) error {
	for {
		if err := handleMsg(admHandler, peer); err != nil {
			peer.Log().Debug("`adamnite` protocol message handler error", "err", err)
			return err
		}
	}
}

func handleMsg(admHandler AdamniteHandlerInterface, peer *Peer) error {
	msg, err := peer.rw.ReadMsg()
	if err != nil {
		return err
	}

	if msg.Size > MaxMsgSize {
		return fmt.Errorf("`adamnite` protocol message too long: %v > %v", msg.Size, MaxMsgSize)
	}

	defer msg.Discard()

	if handler := admProtocolDemo[msg.Code]; handler != nil {
		return handler(admHandler, msg, peer)
	}
	return fmt.Errorf("`adamnite` protocol message invalid `Code`: %v", msg.Code)
}
