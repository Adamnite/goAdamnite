package bargossip

// SubProtocol represents Adamnite blockchain p2p sub protocols that concerned with blockchain.
type SubProtocol struct {
	ProtocolID uint
	Run        func(peer *Peer, rw MsgReader) error
}
