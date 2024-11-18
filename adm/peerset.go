package adm

import (
	"errors"
	"sync"

	"github.com/adamnite/go-adamnite/adm/protocols/adampro"
)

type peerSet struct {
	peers map[string]*adamnitePeer

	lock   sync.RWMutex
	closed bool
}

func newPeerSet() *peerSet {
	return &peerSet{
		peers: make(map[string]*adamnitePeer),
	}
}

func (ps *peerSet) registerPeer(peer *adampro.Peer) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.closed {
		return errors.New("peerset closed")
	}

	id := peer.ID().String()
	if _, ok := ps.peers[id]; ok {
		return errors.New("peer already registered")
	}

	adamPeer := &adamnitePeer{
		Peer: peer,
	}
	ps.peers[id] = adamPeer
	return nil
}
