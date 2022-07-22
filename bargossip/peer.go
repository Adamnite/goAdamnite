// Adamnite remote peer

package bargossip

import (
	"sync"

	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/log15"
)

type Peer struct {
	peerConn *wrapPeerConnection
	log      log15.Logger
	created  mclock.AbsTime

	wg sync.WaitGroup
}

func newPeer(connection *wrapPeerConnection, log log15.Logger, protocols []SubProtocol) *Peer {

}
