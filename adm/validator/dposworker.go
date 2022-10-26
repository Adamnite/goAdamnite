package validator

import (
	"math/big"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
	"github.com/adamnite/go-adamnite/common"

	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/event"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/params"
)

type environment struct {
	signer types.Signer

	state   *statedb.StateDB
	dposEnv *types.DposEnv

	tcount int

	header *types.BlockHeader
	txs    []*types.Transaction
}

type task struct {
	state     *statedb.StateDB
	block     *types.Block
	createdAt time.Time
}

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

	mu      sync.RWMutex
	current *environment

	taskCh chan *task

	snapshotMu    sync.RWMutex //用于保护块快照和状态快照的锁
	snapshotBlock *types.Block
	snapshotState *statedb.StateDB

	fullTaskHook func()
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
	if currentBlock.Numberu64()%dpos.EpochBlockCount == 0 {
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
			CurrentEpoch:    currentBlock.Numberu64() / dpos.EpochBlockCount,
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

func (w *dposWorker) createNewWork() {
	w.mu.RLock()
	defer w.mu.RUnlock()

	tstart := time.Now()
	parent := w.chain.CurrentBlock()

	tstamp := tstart.Unix()
	if parent.Header().Time >= uint64(tstamp) {
		tstamp = int64(parent.Header().Time) + 1
	}

	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		log15.Info("Mining too far in the future", "wait", common.PrettyDuration(wait))
		time.Sleep(wait)
	}

	num := parent.Number()
	header := &types.BlockHeader{
		ParentHash: parent.Hash(),
		Number:     num.Add(num, common.Big1),
		Time:       uint64(tstamp),
	}

	if err := w.dposEngine.Prepare(w.chain, header); err != nil {
		log15.Error("Failed to prepare header for staking", "err", err)
		return
	}

	pending, err := w.adamnite.TxPool().Pending()
	if err != nil {
		log15.Error("Failed to fetch pending transactions", "err", err)
		return
	}

	localTxs, remoteTxs := make(map[common.Address]types.Transactions), pending
	for _, account := range w.adamnite.TxPool().Locals() {
		if txs := remoteTxs[account]; len(txs) > 0 {
			delete(remoteTxs, account)
			localTxs[account] = txs
		}
	}
	// if len(localTxs) > 0 {
	// 	txs := types.NewTransactionsByPriceAndNonce(w.current.signer, localTxs)
	// 	if w.commitTransactions(txs, w.coinbase) {
	// 		return
	// 	}
	// }
	// if len(remoteTxs) > 0 {
	// 	txs := types.NewTransactionsByPriceAndNonce(w.current.signer, remoteTxs)
	// 	if w.commitTransactions(txs, w.coinbase) {
	// 		return
	// 	}
	// }

	err1 := w.commit(w.fullTaskHook, tstart)
	if err1 != nil {
		log15.Error(err1.Error())
		os.Exit(0)
	}
	return
}

func (w *dposWorker) commit(interval func(), start time.Time) error {

	s := w.current.state
	block, err := w.dposEngine.Finalize(w.chain, w.current.header, s, w.current.txs, w.current.dposEnv, *w.adamnite.WitnessCandidatePool())
	if err != nil {
		return err
	}
	block.Header().DposEnv = w.current.header.DposEnv

	w.updateSnapshot()

	return nil
}

func (w *dposWorker) updateSnapshot() {
	w.snapshotMu.Lock()
	defer w.snapshotMu.Unlock()

	w.snapshotBlock = types.NewBlock(
		w.current.header,
		w.current.txs,
		nil,
	)

	w.snapshotState = w.current.state
}

func (w *dposWorker) makeCurrent(parent *types.Block, header *types.BlockHeader) error {
	state, err := w.chain.StateAt(parent.Header().ParentHash)
	if err != nil {
		return err
	}

	trieDB := state.Database().TrieDB()

	dposEnv, err := types.FromProto(trieDB, parent.Header().DposEnv)
	if err != nil {
		return err
	}
	env := &environment{
		signer:  types.AdamniteSigner(*w.chainConfig.ChainID),
		state:   state,
		dposEnv: dposEnv,

		header: header,
	}

	env.tcount = 0
	w.current = env
	return nil
}
