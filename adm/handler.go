package adm

import (
	"sync"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/protocols/adampro"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/event"
	"github.com/adamnite/go-adamnite/log15"
)

type handlerParams struct {
	Database adamnitedb.Database
	TxPool   txPool
	Chain    *blockchain.Blockchain
	ChainID  uint64
	EventMux *event.TypeMux
}

type handler struct {
	chainID uint64

	database adamnitedb.Database
	txpool   txPool
	chain    *blockchain.Blockchain
	maxPeers int

	eventMux    *event.TypeMux
	newBlockSub *event.TypeMuxSubscription

	peers *peerSet

	wg     sync.WaitGroup
	peerWG sync.WaitGroup
}

func newHandler(param *handlerParams) (*handler, error) {
	h := &handler{
		chainID:  param.ChainID,
		database: param.Database,
		txpool:   param.TxPool,
		chain:    param.Chain,
		peers:    newPeerSet(),
		eventMux: param.EventMux,
	}

	return h, nil
}

func (h *handler) Chain() *blockchain.Blockchain { return h.chain }
func (h *handler) TxPool() adampro.TxPool  { return h.txpool }
func (h *handler) RunPeer(peer *adampro.Peer, handler adampro.Handler) error {
	h.peerWG.Add(1)
	defer h.peerWG.Done()

	peer.Log().Debug("Adamnite peer connected", "name", peer.ID())

	if err := h.peers.registerPeer(peer); err != nil {
		peer.Log().Error("Adamnite peer reg failed", "err", err)
		return err
	}

	return handler(peer)
}

func (h *handler) Start(maxPeers int) {
	h.maxPeers = maxPeers

	h.wg.Add(1)
	h.newBlockSub = h.eventMux.Subscribe(blockchain.NewBlockEvent{})
	go h.newBlockBroadcastLoop()
}

func (h *handler) newBlockBroadcastLoop() {
	defer h.wg.Done()

	for block := range h.newBlockSub.Chan() {
		if event, ok := block.Data.(blockchain.NewBlockEvent); ok {
			h.BoradcastBlock(event.Block)
		}
	}
}

func (h *handler) BoradcastBlock(block *types.Block) {
	for _, peer := range h.peers.peers {
		peer.AsyncSendNewBlock(block)
		log15.Debug("Send block to ", "peer", peer.Peer.ID().String(), "block num", block.Numberu64())
	}
}

func (h *handler) Handle(peer *adampro.Peer, packet adampro.Packet) error {
	switch packet := packet.(type) {
	case *adampro.NewBlockPacket:
		return h.handleBlockBroadcast(peer, packet.Block)
	}
	return nil
}

func (h *handler) handleBlockBroadcast(peer *adampro.Peer, block *types.Block) error {
	log15.Info("Import block", "number", block.Numberu64(), "witness", block.Header().Witness.Hex())
	h.chain.AddImportedBlock(block)
	return nil
}
