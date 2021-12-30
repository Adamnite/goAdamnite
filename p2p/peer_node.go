package p2p

// PeerNode holds information about the peer node, lel
type PeerNode struct {
	IP       string `json:"ip"`        // the IP Address of the peer
	Port     uint64 `json:"port"`      // The port the peer is listening on
	IsActive bool   `json:"is-active"` // Indicates whether the peer is active or not
}

type DBHash string // type holder for the hash
type StatusRes struct {
	Hash       DBHash
	Number     uint64
	KnownPeers []PeerNode
}
