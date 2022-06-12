package adm

import (
	"sync"

	"github.com/adamnite/go-adamnite/adm/protocols/adampro"
)

type adamnitePeerInfo struct {
	Version uint   `json:"version"`
	Head    string `json:"head"`
}

type adamnitePeer struct {
	*adampro.Peer

	lock sync.RWMutex
}

func (p *adamnitePeer) info() *adamnitePeerInfo {
	hash := p.Head()

	return &adamnitePeerInfo{
		Version: p.Version(),
		Head:    hash.Hex(),
	}
}
