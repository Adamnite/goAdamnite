package adampro

import (
	"fmt"

	"github.com/adamnite/go-adamnite/p2p"
	"github.com/adamnite/go-adamnite/p2p/enode"
	"github.com/adamnite/go-adamnite/p2p/enr"
)

type Handler func(peer *Peer) error
type msgHandler func(admHandler AdamniteHandlerInterface, msg Decoder, peer *Peer) error

var admProtocolDemo = map[uint64]msgHandler{
	NewBlockMsg: newBlockHandler,
}

func MakeProtocols(adamniteHandler AdamniteHandlerInterface, chainID uint64, dnsdisc enode.Iterator) []p2p.Protocol {
	protocols := make([]p2p.Protocol, len(ProtocolVersions))

	for i, v := range ProtocolVersions {
		protocols[i] = p2p.Protocol{
			Name:    ProtocolName,
			Version: v,
			Length:  ProtocolLength[v],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				peer := NewPeer(v, p, rw, adamniteHandler.TxPool())
				defer peer.Close()
				return adamniteHandler.RunPeer(peer, func(peer *Peer) error {
					return ProtocolMsgHandler(adamniteHandler, peer)
				})
			},
			PeerInfo: func(id enode.ID) interface{} {
				return nil
			},
			Attributes: []enr.Entry{},
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
