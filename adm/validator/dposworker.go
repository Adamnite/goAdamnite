package validator

import (
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/utils"

	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/event"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/params"
)

type environment struct {
	signer types.Signer
	state  *statedb.StateDB
	tcount int
	header *types.BlockHeader
	txs    []*types.Transaction
}

type task struct {
	state     *statedb.StateDB
	block     *types.Block
	createdAt time.Time
}

const (
	commitInterruptNone int32 = iota
	commitInterruptNewHead
	commitInterruptResubmit

	resultQueueSize = 10

	txChanSize = 4096

	chainHeadChanSize = 10

	chainSideChanSize = 10

	resubmitAdjustChanSize = 10

	minRecommitInterval = 1 * time.Second

	maxRecommitInterval = 15 * time.Second

	intervalAdjustRatio = 0.1

	intervalAdjustBias = 200 * 1000.0 * 1000.0

	staleThreshold = 7
)

// newWorkReq represents a request for new  work submitting with relative interrupt notifier.
type newWorkReq struct {
	interrupt *int32
	noempty   bool
	timestamp int64
}

// intervalAdjust represents a resubmitting interval adjustment.
type intervalAdjust struct {
	ratio float64
	inc   bool
}

type dposWorker struct {
	config      *Config
	chainConfig *params.ChainConfig
	dposEngine  dpos.DPOS
	adamnite    AdamniteImplInterface
	chain       *core.Blockchain

	mux            *event.TypeMux
	importBlockCh  chan core.ImportBlockEvent
	importBlockSub event.Subscription

	txsCh        chan core.NewTxsEvent
	txsSub       event.Subscription
	chainHeadCh  chan core.ChainHeadEvent
	chainHeadSub event.Subscription

	// channels
	newWorkCh          chan *newWorkReq
	taskCh             chan *task
	startCh            chan struct{}
	exitCh             chan struct{}
	resultCh           chan *types.Block
	resubmitIntervalCh chan time.Duration
	resubmitAdjustCh   chan *intervalAdjust
	genBlockCh         chan *types.Block

	running int32 // The status of the DPOS engine

	mu            sync.RWMutex
	current       *environment
	pendingMu     sync.RWMutex
	pendingTasks  map[utils.Hash]*task
	snapshotMu    sync.RWMutex
	snapshotBlock *types.Block
	snapshotState *statedb.StateDB

	coinbase utils.Address
	extra    []byte
	newTxs   int32 // Count of newly arrived transactions since the last seal job commit.

	fullTaskHook func() // Method to call before pushing the full sealing task.
	resubmitHook func(time.Duration, time.Duration)
}

func newDposWorker(config *Config, chainConfig *params.ChainConfig, dpos dpos.DPOS, adamnite AdamniteImplInterface, mux *event.TypeMux, init bool) *dposWorker {
	worker := &dposWorker{
		config:      config,
		chainConfig: chainConfig,
		dposEngine:  dpos,
		adamnite:    adamnite,
		chain:       adamnite.Blockchain(),

		mux:                mux,
		txsCh:              make(chan core.NewTxsEvent, txChanSize),
		chainHeadCh:        make(chan core.ChainHeadEvent, chainHeadChanSize),
		newWorkCh:          make(chan *newWorkReq),
		taskCh:             make(chan *task),
		startCh:            make(chan struct{}),
		exitCh:             make(chan struct{}),
		genBlockCh:         make(chan *types.Block),
		resultCh:           make(chan *types.Block, resultQueueSize),
		importBlockCh:      make(chan core.ImportBlockEvent),
		resubmitIntervalCh: make(chan time.Duration),
		resubmitAdjustCh:   make(chan *intervalAdjust, 10),
	}

	worker.importBlockSub = adamnite.Blockchain().SubscribeImportBlockEvent(worker.importBlockCh)
	worker.txsSub = adamnite.TxPool().SubscribeNewTxsEvent(worker.txsCh)
	worker.chainHeadSub = adamnite.Blockchain().SubscribeChainHeadEvent(worker.chainHeadCh)
	recommit := worker.config.Recommit

	go worker.mainLoop()
	go worker.genBlockLoop()
	go worker.newWorkLoop(recommit)

	if init {
		worker.startCh <- struct{}{}
	}
	return worker
}

func (w *dposWorker) setCoinbase(addr utils.Address) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.coinbase = addr
}

func (w *dposWorker) setExtra(extra []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.extra = extra
}

func (w *dposWorker) setRecommitInterval(interval time.Duration) {
	w.resubmitIntervalCh <- interval
}

func (w *dposWorker) start() {
	atomic.StoreInt32(&w.running, 1)
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

func (w *dposWorker) newWorkLoop(recommit time.Duration) {
	var (
		interrupt   *int32
		minRecommit = recommit // minimal resubmit interval specified by user.
		timestamp   int64      // timestamp for each round of mining.
	)

	timer := time.NewTimer(0)
	<-timer.C // discard the initial tick

	// commit aborts in-flight transaction execution with given signal and resubmits a new one.
	commit := func(noempty bool, s int32) {
		if interrupt != nil {
			atomic.StoreInt32(interrupt, s)
		}
		interrupt = new(int32)
		w.newWorkCh <- &newWorkReq{interrupt: interrupt, noempty: noempty, timestamp: timestamp}
		timer.Reset(recommit)
		atomic.StoreInt32(&w.newTxs, 0)
	}
	// recalcRecommit recalculates the resubmitting interval upon feedback.
	recalcRecommit := func(target float64, inc bool) {
		var (
			prev = float64(recommit.Nanoseconds())
			next float64
		)
		if inc {
			next = prev*(1-intervalAdjustRatio) + intervalAdjustRatio*(target+intervalAdjustBias)
			// Recap if interval is larger than the maximum time interval
			if next > float64(maxRecommitInterval.Nanoseconds()) {
				next = float64(maxRecommitInterval.Nanoseconds())
			}
		} else {
			next = prev*(1-intervalAdjustRatio) + intervalAdjustRatio*(target-intervalAdjustBias)
			// Recap if interval is less than the user specified minimum
			if next < float64(minRecommit.Nanoseconds()) {
				next = float64(minRecommit.Nanoseconds())
			}
		}
		recommit = time.Duration(int64(next))
	}
	// clearPending cleans the stale pending tasks.
	clearPending := func(number uint64) {
		w.pendingMu.Lock()
		for h, t := range w.pendingTasks {
			if t.block.Numberu64()+staleThreshold <= number {
				delete(w.pendingTasks, h)
			}
		}
		w.pendingMu.Unlock()
	}

	for {
		select {
		case <-w.startCh:
			clearPending(w.chain.CurrentBlock().Numberu64())
			timestamp = time.Now().Unix()
			commit(false, commitInterruptNewHead)

		case head := <-w.chainHeadCh:
			clearPending(head.Block.Numberu64())
			timestamp = time.Now().Unix()
			commit(false, commitInterruptNewHead)

		case <-timer.C:
			// If staking is running resubmit a new work cycle periodically to pull in
			// higher priced transactions. Disable this overhead for pending blocks.
			if w.isRunning() {
				// Short circuit if no new transaction arrives.
				if atomic.LoadInt32(&w.newTxs) == 0 {
					timer.Reset(recommit)
					continue
				}
				commit(true, commitInterruptResubmit)
			}

		case interval := <-w.resubmitIntervalCh:
			// Adjust resubmit interval explicitly by user.
			if interval < minRecommitInterval {
				log15.Warn("Sanitizing miner recommit interval", "provided", interval, "updated", minRecommitInterval)
				interval = minRecommitInterval
			}
			log15.Info("Miner recommit interval update", "from", minRecommit, "to", interval)
			minRecommit, recommit = interval, interval

			if w.resubmitHook != nil {
				w.resubmitHook(minRecommit, recommit)
			}

		case adjust := <-w.resubmitAdjustCh:
			// Adjust resubmit interval by feedback.
			if adjust.inc {
				before := recommit
				recalcRecommit(float64(recommit.Nanoseconds())/adjust.ratio, true)
				log15.Info("Increase staker recommit interval", "from", before, "to", recommit)
			} else {
				before := recommit
				recalcRecommit(float64(minRecommit.Nanoseconds()), false)
				log15.Info("Decrease staker recommit interval", "from", before, "to", recommit)
			}

			if w.resubmitHook != nil {
				w.resubmitHook(minRecommit, recommit)
			}

		case <-w.exitCh:
			return
		}
	}
}

func (w *dposWorker) mainLoop() {
	defer w.txsSub.Unsubscribe()

	for {
		select {
		case req := <-w.newWorkCh:
			w.createNewWork(req.interrupt, req.noempty, req.timestamp)

		case ev := <-w.txsCh:

			if !w.isRunning() && w.current != nil {
				// If block is already full, abort

				w.mu.RLock()
				coinbase := w.coinbase
				w.mu.RUnlock()

				txs := make(map[utils.Address]types.Transactions)
				for _, tx := range ev.Txs {
					acc, _ := types.Sender(w.current.signer, tx)
					txs[acc] = append(txs[acc], tx)
				}
				txset := types.NewTransactionsByPriceAndNonce(w.current.signer, txs)
				tcount := w.current.tcount
				w.commitTransactions(txset, coinbase)
				// Only update the snapshot if any new transactons were added
				// to the pending block
				if tcount != w.current.tcount {
					w.updateSnapshot()
				}
			}
			atomic.AddInt32(&w.newTxs, int32(len(ev.Txs)))

		// System stopped
		case <-w.exitCh:
			return
		case <-w.txsSub.Err():
			return

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

func (w *dposWorker) createNewWork(interrupt *int32, noempty bool, timestamp int64) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	tstart := time.Now()
	parent := w.chain.CurrentBlock()

	// var wAddr utils.Address
	// if parent.Numberu64()%dpos.EpochBlockCount == 0 {
	// 	wAddr = w.adamnite.WitnessPool().GetCurrentWitnessAddress(nil)
	// } else {
	// 	wAddr = w.adamnite.WitnessPool().GetCurrentWitnessAddress(&parent.Header().Witness)
	// }

	tstamp := tstart.Unix()
	if parent.Header().Time >= uint64(tstamp) {
		tstamp = int64(parent.Header().Time) + 1
	}

	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		log15.Info("too far in the future", "wait", utils.PrettyDuration(wait))
		time.Sleep(wait)
	}

	num := parent.Number()
	header := &types.BlockHeader{
		ParentHash:      parent.Hash(),
		Time:            uint64(time.Now().Unix()),
		WitnessRoot:     utils.HexToHash("0x00000000000"),
		Number:          num.Add(num, utils.Big1),
		Signature:       utils.HexToHash("0x00000"),
		TransactionRoot: utils.HexToHash("0x0000"),
		CurrentEpoch:    parent.Numberu64() / dpos.EpochBlockCount,
		StateRoot:       utils.HexToHash("0x0000"),
		Extra:           w.extra,
	}

	if err := w.dposEngine.Prepare(w.chain, header); err != nil {
		// log15.Error("Failed to prepare header for staking", "err", err)
		return
	}

	err := w.makeCurrent(parent, header)
	if err != nil {
		log15.Error("Failed to create mining context", "err", err)
		return
	}
	pending, err := w.adamnite.TxPool().Pending()
	if err != nil {
		log15.Error("Failed to fetch pending transactions", "err", err)
		return
	}

	localTxs, remoteTxs := make(map[utils.Address]types.Transactions), pending
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

func (w *dposWorker) commitTransaction(tx *types.Transaction, coinbase utils.Address) error {

	w.current.txs = append(w.current.txs, tx)

	return nil
}

func (w *dposWorker) commitTransactions(txs *types.TransactionsByPriceAndNonce, coinbase utils.Address) bool {

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

		w.current.state.Prepare(tx.Hash(), utils.Hash{}, w.current.tcount)

		err := w.commitTransaction(tx, coinbase)
		if err != nil {
			log15.Error("commit transation failed", err.Error())
		}

	}

	return false
}
func (w *dposWorker) commit(interval func(), start time.Time) error {

	s := w.current.state.Copy()
	_, err := w.dposEngine.Finalize(w.chain, w.current.header, s, w.current.txs)
	if err != nil {
		return err
	}

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

	if err != nil {
		return err
	}
	env := &environment{
		signer: types.AdamniteSigner{},
		state:  state,

		header: header,
	}

	env.tcount = 0
	w.current = env
	return nil
}
