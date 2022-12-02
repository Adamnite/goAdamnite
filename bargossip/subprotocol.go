package bargossip

// SubProtocol represents Adamnite blockchain p2p sub protocols that concerned with blockchain.
type SubProtocol struct {
	ProtocolID         uint
	ProtocolCodeOffset uint
	ProtocolLength     uint
	Run                func(peer *Peer, rw MsgReadWriter) error
}
