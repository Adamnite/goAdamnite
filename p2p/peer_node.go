package p2p

// PeerNode holds information about the peer node, lel
type PeerNode struct {
	IP          string // the IP Address of the peer
	Port        uint64 // The port the peer is listening on
	IsBootStrap bool   // IsBootStrap, indicates whether the node is bootstrap node or not
	IsActive    bool   // Indicates whether the peer is active or not
}

type DBHash string // type holder for the hash
type StatusRes struct {
	Hash       DBHash
	Number     uint64
	KnownPeers []PeerNode
}
