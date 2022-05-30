package p2p

import (
	"fmt"

	"github.com/adamnite/go-adamnite/p2p/enode"
	"github.com/adamnite/go-adamnite/p2p/enr"
)

type PeerCapability struct {
	Name    string
	Version uint
}

func (cap PeerCapability) String() string {
	return fmt.Sprintf("%s/%d", cap.Name, cap.Version)
}

type Protocol struct {
	Name           string
	Version        uint
	Length         uint64
	Run            func(peer *Peer, rw MsgReadWriter) error
	PeerInfo       func(id enode.ID) interface{}
	DialCandidates enode.Iterator
	Attributes     []enr.Entry
}

func (p Protocol) cap() PeerCapability {
	return PeerCapability{p.Name, p.Version}
}

type capsByNameAndVersion []PeerCapability

func (cs capsByNameAndVersion) Len() int      { return len(cs) }
func (cs capsByNameAndVersion) Swap(i, j int) { cs[i], cs[j] = cs[j], cs[i] }
func (cs capsByNameAndVersion) Less(i, j int) bool {
	return cs[i].Name < cs[j].Name || (cs[i].Name == cs[j].Name && cs[i].Version < cs[j].Version)
}
