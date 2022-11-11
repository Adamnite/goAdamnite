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

	snapshotMu    sync.RWMutex
	snapshotBlock *types.Block
	snapshotState *statedb.StateDB

	coinbase common.Address

	newTxs       int32 // Count of newly arrived transactions since the last seal job commit.
	quitCh       chan struct{}
	stopper      chan struct{}
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

func (w *dposWorker) setCoinbase(addr common.Address) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.coinbase = addr
}

func (w *dposWorker) start() {
	atomic.StoreInt32(&w.running, 1)
	go w.mintLoop(16)
	w.startCh <- struct{}{}
}

//Pending returns the Pending state and the corresponding block.
func (w *dposWorker) pending() (*types.Block, *statedb.StateDB) {
	//return snapshot to avoid contention on current mutex
	w.snapshotMu.RLock()
	defer w.snapshotMu.RUnlock()
	if w.snapshotState == nil {
		return nil, nil
	}
	return w.snapshotBlock, w.snapshotState.Copy()
}

//PendingBlock returns PendingBlock.
func (w *dposWorker) pendingBlock() *types.Block {
	//return snapshot to avoid contention on current mutex
	w.snapshotMu.RLock()
	defer w.snapshotMu.RUnlock()
	return w.snapshotBlock
}

func (w *dposWorker) mintBlock(now int64, blockInterval uint64) {
	engine, ok := w.dposEngine.(*dpos.AdamniteDPOS)
	if !ok {
		log15.Error("Only the dpos engine was allowed")
		return
	}
	//Check whether the current validator is the current node
	err := engine.CheckValidator(w.chain.CurrentBlock(), now, blockInterval)
	if err != nil {
		switch err {
		case dpos.ErrWaitForPrevBlock,
			dpos.ErrApplyNextBlock,
			dpos.ErrInvalidApplyBlockTime,
			dpos.ErrInvalidWitness:
			log15.Debug("Failed to stake the block, while ", "err", err)
		default:
			log15.Error("Failed to stake the block", "err", err)
		}
		return
	}
	w.createNewWork()

}

func (w *dposWorker) mintLoop(blockInterval uint64) {
	wt := time.Duration(int64(blockInterval))
	//The default wt is "time.second", accounting blockinterval gets the waiting time
	ticker := time.NewTicker(wt * time.Second / 10).C // Chanel
	for {
		select {
		case now := <-ticker:
			atomic.StoreInt32(&w.newTxs, 0)
			w.mintBlock(now.Unix(), blockInterval)
		case <-w.stopper:
			close(w.quitCh)
			w.quitCh = make(chan struct{}, 1)
			w.stopper = make(chan struct{}, 1)
			return
		}
	}
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

	var wAddr common.Address
	if parent.Numberu64()%dpos.EpochBlockCount == 0 {
		wAddr = w.adamnite.WitnessPool().GetCurrentWitnessAddress(nil)
	} else {
		wAddr = w.adamnite.WitnessPool().GetCurrentWitnessAddress(&parent.Header().Witness)
	}

	tstamp := tstart.Unix()
	if parent.Header().Time >= uint64(tstamp) {
		tstamp = int64(parent.Header().Time) + 1
	}

	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		log15.Info("too far in the future", "wait", common.PrettyDuration(wait))
		time.Sleep(wait)
	}

	num := parent.Number()
	header := &types.BlockHeader{
		ParentHash:      parent.Hash(),
		Time:            uint64(time.Now().Unix()),
		Witness:         wAddr,
		WitnessRoot:     common.HexToHash("0x00000000000"),
		Number:          num,
		Signature:       common.HexToHash("0x00000"),
		TransactionRoot: common.HexToHash("0x0000"),
		CurrentEpoch:    parent.Numberu64() / dpos.EpochBlockCount,
		StateRoot:       common.HexToHash("0x0000"),
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
	if len(localTxs) > 0 {
		txs := types.NewTransactionsByPriceAndNonce(w.current.signer, localTxs)
		if w.commitTransactions(txs, w.coinbase) {
			return
		}
	}
	if len(remoteTxs) > 0 {
		txs := types.NewTransactionsByPriceAndNonce(w.current.signer, remoteTxs)
		if w.commitTransactions(txs, w.coinbase) {
			return
		}
	}

	err1 := w.commit(w.fullTaskHook, tstart)
	if err1 != nil {
		log15.Error(err1.Error())
		os.Exit(0)
	}
	return
}

func (w *dposWorker) commitTransaction(tx *types.Transaction, coinbase common.Address) error {

	w.current.txs = append(w.current.txs, tx)

	return nil
}

func (w *dposWorker) commitTransactions(txs *types.TransactionsByPriceAndNonce, coinbase common.Address) bool {

	if w.current == nil {
		return true
	}

	for {
		// Retrieve the next transaction and abort if all done
		tx := txs.Peek()
		if tx == nil {
			break
		}
		types.Sender(w.current.signer, tx)

		w.current.state.Prepare(tx.Hash(), common.Hash{}, w.current.tcount)

		err := w.commitTransaction(tx, coinbase)
		if err != nil {
			log15.Error("commit transation failed", err.Error())
		}

	}

	return false
}
func (w *dposWorker) commit(interval func(), start time.Time) error {

	s := w.current.state
	block, err := w.dposEngine.Finalize(w.chain, w.current.header, s, w.current.txs, *w.current.dposEnv, *w.adamnite.WitnessCandidatePool())
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

	dposEnv, err := types.NewDposEnvFromProto(trieDB, &(parent.Header().DposEnv))
	if err != nil {
		return err
	}
	env := &environment{
		signer:  types.AdamniteSigner{},
		state:   state,
		dposEnv: dposEnv,

		header: header,
	}

	env.tcount = 0
	w.current = env
	return nil
}
