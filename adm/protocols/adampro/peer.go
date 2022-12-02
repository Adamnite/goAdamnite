package adampro

import (
	"sync"

	"github.com/adamnite/go-adamnite/bargossip"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"

	gset "github.com/deckarep/golang-set"
	mapset "github.com/deckarep/golang-set"
)

const (
	maxQueuedBlockAnns = 4
	maxQueuedBlocks    = 4

	maxKnownBlocks = 1024
)

type Peer struct {
	id string
	*bargossip.Peer
	rw      bargossip.MsgReadWriter
	version uint

	blockHead common.Hash // head block hash

	knownBlocks         gset.Set
	queuedBlocks        chan *types.Block
	anounceQueuedBlocks chan *types.Block

	txpool      TxPool
	knownTxs    gset.Set
	txBroadcast chan []common.Hash
	txAnnounce  chan []common.Hash

	terminate chan struct{}
	lock      sync.RWMutex
}

func NewPeer(version uint, p *bargossip.Peer, rw bargossip.MsgReadWriter, txpool TxPool) *Peer {
	peer := &Peer{
		id:                  p.ID().String(),
		Peer:                p,
		rw:                  rw,
		version:             version,
		knownTxs:            mapset.NewSet(),
		knownBlocks:         mapset.NewSet(),
		queuedBlocks:        make(chan *types.Block, maxQueuedBlocks),
		anounceQueuedBlocks: make(chan *types.Block, maxQueuedBlockAnns),
		txBroadcast:         make(chan []common.Hash),
		txAnnounce:          make(chan []common.Hash),
		txpool:              txpool,
		terminate:           make(chan struct{}),
	}

	go peer.broadcastBlocks()

	return peer
}

func (p *Peer) Head() (hash common.Hash) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	copy(hash[:], p.blockHead[:])
	return hash
}

func (p *Peer) Version() uint {
	return p.version
}

func (p *Peer) Close() {
	close(p.terminate)
}

func (p *Peer) AsyncSendNewBlock(block *types.Block) {
	select {
	case p.queuedBlocks <- block:
	default:
		p.Log().Debug("Dropping block", "number", block.Numberu64(), "witness", block.Header().Witness.Hex())
	}
}

func (p *Peer) broadcastBlocks() {
	for {
		select {
		case block := <-p.queuedBlocks:
			if err := p.SendNewBlock(block); err != nil {
				return
			}
		}
	}
}

func (p *Peer) SendNewBlock(block *types.Block) error {
	for p.knownBlocks.Cardinality() >= maxKnownBlocks {
		p.knownBlocks.Pop()
	}
	p.knownBlocks.Add(block.Hash())

	return bargossip.Send(p.rw, NewBlockMsg, &NewBlockPacket{
		Block: block,
	})
}
