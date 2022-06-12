package validator

import (
	"math/big"
	"sync/atomic"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/dpos/adamdpos"
	"github.com/adamnite/go-adamnite/event"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/params"
)

type dposWorker struct {
	config      *Config
	chainConfig *params.ChainConfig
	dposEngine  dpos.DPOS
	adamnite    AdamniteImplInterface
	chain       *core.Blockchain

	mux *event.TypeMux

	// channels
	startCh        chan struct{}
	exitCh         chan struct{}
	genBlockCh     chan *types.Block
	importBlockCh  chan core.ImportBlockEvent
	importBlockSub event.Subscription

	running int32 // The status of the DPOS engine
}

func newDposWorker(config *Config, chainConfig *params.ChainConfig, dpos dpos.DPOS, adamnite AdamniteImplInterface, mux *event.TypeMux) *dposWorker {
	worker := &dposWorker{
		config:      config,
		chainConfig: chainConfig,
		dposEngine:  dpos,
		adamnite:    adamnite,
		chain:       adamnite.Blockchain(),

		mux: mux,

		startCh:       make(chan struct{}),
		exitCh:        make(chan struct{}),
		genBlockCh:    make(chan *types.Block),
		importBlockCh: make(chan core.ImportBlockEvent),
	}

	worker.importBlockSub = adamnite.Blockchain().SubscribeImportBlockEvent(worker.importBlockCh)

	go worker.mainLoop()
	go worker.genBlockLoop()
	return worker
}

func (w *dposWorker) start() {
	atomic.StoreInt32(&w.running, 1)
	w.startCh <- struct{}{}
}

func (w *dposWorker) stop() {
	atomic.StoreInt32(&w.running, 0)
}

func (w *dposWorker) isRunning() bool {
	return atomic.LoadInt32(&w.running) == 1
}

// close terminates all background threads maintained by the worker.
// Note the worker does not support being closed multiple times.
func (w *dposWorker) close() {
	atomic.StoreInt32(&w.running, 0)
	close(w.exitCh)
}

func (w *dposWorker) mainLoop() {
	for {
		select {
		case <-w.exitCh:
			return
		case <-w.startCh:
			go w.commitBlock()
		case <-w.importBlockCh:
			go w.commitBlock()
		}
	}
}

func (w *dposWorker) commitBlock() {
	currentBlock := w.adamnite.Blockchain().CurrentBlock()
	var wAddr common.Address
	if currentBlock.Numberu64()%adamdpos.EpochBlockCount == 0 {
		wAddr = w.adamnite.WitnessPool().GetCurrentWitnessAddress(nil)
	} else {
		wAddr = w.adamnite.WitnessPool().GetCurrentWitnessAddress(&currentBlock.Header().Witness)
	}

	if wAddr == w.config.WitnessAddress {
		newBlockHeader := &types.BlockHeader{
			ParentHash:      currentBlock.Hash(),
			Time:            uint64(time.Now().Unix()),
			Witness:         wAddr,
			WitnessRoot:     common.HexToHash("0x00000000000"),
			Number:          new(big.Int).Add(currentBlock.Number(), big.NewInt(1)),
			Signature:       common.HexToHash("0x00000"),
			TransactionRoot: common.HexToHash("0x0000"),
			CurrentEpoch:    currentBlock.Numberu64() / adamdpos.EpochBlockCount,
			StateRoot:       common.HexToHash("0x0000"),
		}

		t := time.NewTimer(15 * time.Second)
		defer t.Stop()
		select {
		case <-t.C:
			block := types.NewBlock(newBlockHeader, nil, trie.NewStackTrie(nil))
			w.genBlockCh <- block
		}

	}
}

func (w *dposWorker) genBlockLoop() {
	for {
		select {
		case block := <-w.genBlockCh:
			err := w.chain.WriteBlock(block)
			if err != nil {
				log15.Error("Failed writing block to chain", "err", err)
				continue
			}

			w.mux.Post(core.NewBlockEvent{Block: block})

			log15.Info("Created new block", "number", block.Number(), "witness", block.Header().Witness)
		}
	}
}
