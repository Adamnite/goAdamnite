package pendingHandling

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/safe"
)

type processSteps int8

const (
	waitingToStart processSteps = iota
	waitingOnParent
	waitingOnPeers
	running
	failed
	completed
)

type complexTransaction struct {
	lock           sync.RWMutex
	transaction    *utils.VMCallTransaction
	needsToBeAfter *complexTransaction //eg, if someone sends a balance more than once, the first one should always be the one to go through
	needsToRunWith []*complexTransaction
	isPlaceholder  bool //if true, this is a placeholder on another contract to be called
	finalChanges   *utils.RuntimeChanges
	processStep    chan processSteps //set to true when it starts, and false when completed
}

func newComplexTransaction(t *utils.VMCallTransaction, firstNeeds *complexTransaction) (*complexTransaction, error) {
	//here is where we would process if anything needs to be before it.
	ct := complexTransaction{
		transaction:    t,
		needsToBeAfter: firstNeeds,
		processStep:    make(chan processSteps, waitingToStart),
		isPlaceholder:  false,
	}
	//TODO: check if it calls any other contracts. If it does, add those to "needs to run with"
	return &ct, nil

}

// run this transaction (and any required transactions) on this VM instance
func (ct *complexTransaction) RunOn(vm *VM.Machine) (hash []byte, err error) {
	ct.lock.Lock()
	defer ct.lock.Unlock()
	if ct.needsToBeAfter != nil {
		ct.processStep <- waitingOnParent
		//if we have anything that *must* run before this, we wait
		step := <-ct.needsToBeAfter.processStep
		for step != completed {
			if step == failed {
				ct.processStep <- failed
				return nil, fmt.Errorf("failed because required parent failed")
			}
			step = <-ct.needsToBeAfter.processStep
		}
	}
	if len(ct.needsToRunWith) != 0 {
		//means this contract calls another contract
		ct.processStep <- waitingOnParent
	}
	for _, peer := range ct.needsToRunWith {
		step := <-peer.processStep
		for step != waitingOnPeers {
			if step == failed {
				go func() { ct.processStep <- failed }()
				return nil, fmt.Errorf("failed because required peer could not load")
			}
		}
	}
	go func() { ct.processStep <- running }()
	ans, err := vm.CallOnContractWith(ct.transaction.VMInteractions)
	if err != nil {
		go func() { ct.processStep <- failed }()
		return nil, err
	}
	ct.finalChanges = ans
	go func() { ct.processStep <- completed }()
	return ans.Hash().Bytes(), nil
}

type contractHeld struct {
	// contractCalled common.Address
	lock                 sync.Mutex            //only one thread is to touch this at a time!
	vm                   *VM.Machine           //the vm with all of these changes
	transactions         []*complexTransaction //the transactions entered
	nextToRun            int                   //index of the next transaction to process
	runningHash          []byte
	successfullyRunCount *safe.SafeInt
}

func newContractHeld(contract *common.Address, db VM.DBInterfaceItem) (*contractHeld, error) {
	vm, err := VM.NewVirtualMachineWithDB(contract, db)
	if err != nil {
		return nil, err
	}
	ch := contractHeld{
		vm:                   vm,
		transactions:         []*complexTransaction{},
		nextToRun:            0,
		runningHash:          vm.GetContractHash().Bytes(),
		successfullyRunCount: safe.NewSafeInt(0),
	}
	return &ch, nil
}

// used to safely get the transaction count. Locks the contract held while doing so!
func (ch *contractHeld) getCurrentAndMaxTransaction() (current int, max int) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	return ch.nextToRun, len(ch.transactions)
}

// step to the next transaction
func (ch *contractHeld) Step(state *statedb.StateDB, addSuccessTo *safe.SafeSlice) error {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	if ch.nextToRun >= len(ch.transactions) {
		return fmt.Errorf("out of transactions to run. Current next up %d of %d", ch.nextToRun, len(ch.transactions))
	}
	//TODO: you should move the sent funds here
	hash, err := ch.transactions[ch.nextToRun].RunOn(ch.vm)
	//TODO: you should spend their gas here

	//add the running hash that was given to this transaction before its calling.
	ch.transactions[ch.nextToRun].transaction.RunnerHash = ch.runningHash
	ch.nextToRun += 1
	if err != nil {
		//TODO: here should undo transfers if it the transaction failed.
		//TODO: we might need to also revert the gas fees spent... since failed transactions aren't added to the block...
		return err
	}
	if addSuccessTo != nil {
		addSuccessTo.Append(ch.transactions[ch.nextToRun-1].transaction)
	}
	ch.successfullyRunCount.Add(1)
	ch.runningHash = crypto.Sha512(append(ch.runningHash, hash...))
	return nil
}

func (ch *contractHeld) RunAll(state *statedb.StateDB) error {
	//do this weirdness so that we can get the current point and the endpoint, while locking up the contract held as little as possible
	for currentPoint, endPoint := ch.getCurrentAndMaxTransaction(); currentPoint < endPoint; currentPoint, endPoint = ch.getCurrentAndMaxTransaction() {
		if err := ch.Step(state, nil); err != nil {
			return err
		}
	}
	return nil
}
func (ch *contractHeld) AddTransactionToQueue(ct *complexTransaction) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	ch.transactions = append(ch.transactions, ct)
}

type ContractStateHolder struct {
	lock                 sync.Mutex
	contractsHeld        map[string]*contractHeld         //contractAddressString(hex)->contractHeld
	sends                map[string][]*complexTransaction //from(hexOfAddress)->all instances they have sent to, in order
	receives             map[string][]*complexTransaction //from(hexOfAddress)->all instances in which  they receive nite
	dbCache              VM.DBCacheAble                   //for keeping a local reference. allows each call to be made once
	successfullyRunCount *safe.SafeInt
}

func NewContractStateHolder(dbCache VM.DBCacheAble) (*ContractStateHolder, error) {
	csh := ContractStateHolder{
		contractsHeld:        map[string]*contractHeld{},
		sends:                map[string][]*complexTransaction{},
		receives:             map[string][]*complexTransaction{},
		dbCache:              dbCache,
		successfullyRunCount: safe.NewSafeInt(0),
	}
	return &csh, nil
}
func (csh *ContractStateHolder) QueueTransaction(t utils.TransactionType) (err error) {
	//assume that the transaction is indeed, intended for us to handle it
	//also assume that the transactions are being fed to us in order
	csh.lock.Lock()
	defer csh.lock.Unlock()
	//check if the target contract is already loaded
	if t.GetType() == utils.Transaction_Basic {
		return fmt.Errorf("expected VM calling type, got base transaction type")
	}
	if t.GetType() == utils.Transaction_VM_NewContract {
		//TODO: we need to handle contract creation
		return nil
	}
	vmT := t.(*utils.VMCallTransaction).VMInteractions
	ch, exists := csh.contractsHeld[vmT.ContractCalled.Hex()]
	if !exists {
		//it's not stored yet, so we need to make this
		ch, err = newContractHeld(&vmT.ContractCalled, csh.dbCache)
		if err != nil {
			return err
		}
		go func() {
			//this is to update how many things we've processed since starting
			runchan := ch.successfullyRunCount.GetOnUpdate()
			for {
				updatedVal, ok := <-runchan
				if !ok {
					return
				}
				if updatedVal != 0 {
					csh.successfullyRunCount.Add(1)
				}
			}
		}()
		csh.contractsHeld[vmT.ContractCalled.Hex()] = ch
	}

	//then check if the target method is loaded
	if err := csh.dbCache.PreCacheCode(vmT.ParametersPassed[:utils.FunctionIdentifierLength]); err != nil {
		return err
	}
	//TODO: check if the method calls another contract. If it does, we need to add a special point in that contracts queue
	var sendsMoney bool = t.(*utils.VMCallTransaction).Amount.Cmp(big.NewInt(0)) != 0
	var sendingBefore *complexTransaction = nil
	//then we actually log the transaction (so we know it'll work)
	if sendsMoney {
		//they send money... so we need to check if they're sending before. Or if they're receiving before
		from := t.(*utils.VMCallTransaction).From.GetAddress().Hex()
		if previouslySent, exists := csh.sends[from]; exists {
			sendingBefore = previouslySent[len(previouslySent)-1]
		} else if previouslyReceived, exists := csh.receives[from]; exists {
			sendingBefore = previouslyReceived[len(previouslyReceived)-1]
		}
	}
	ct, err := newComplexTransaction(t.(*utils.VMCallTransaction), sendingBefore)
	if err != nil {
		return err
	}
	ch.AddTransactionToQueue(ct)
	if sendsMoney {
		//we have to add that they sent nite, with reference to when they do, so it can be added.
		from := t.(*utils.VMCallTransaction).From.GetAddress().Hex()
		if previouslySent, exists := csh.sends[from]; exists {
			csh.sends[from] = append(previouslySent, ct)
		} else {
			csh.sends[from] = []*complexTransaction{ct}
		}

		to := t.(*utils.VMCallTransaction).To.Hex()
		if previouslyReceived, exists := csh.receives[to]; exists {
			csh.receives[to] = append(previouslyReceived, ct)
		} else {
			csh.receives[from] = []*complexTransaction{ct}
		}
	}
	return nil
}

func (csh *ContractStateHolder) RunAll(state *statedb.StateDB) (err error) {
	csh.lock.Lock()
	defer csh.lock.Unlock()
	steppingErr := make(chan error)
	processedCount := safe.NewSafeInt(0)
	go func() {
		//this thread is setup to handle if we complete everything
		for processedCount.Get() < len(csh.contractsHeld) {
			time.After(time.Nanosecond)
		}
		steppingErr <- nil //just send something so that it's marked that we're done

	}()
	for _, current := range csh.contractsHeld {
		go func(ch *contractHeld) {
			if runErr := ch.RunAll(state); runErr != nil {
				//it had a runtime error. In future, we will see if we can try starting it again.
				//TODO: a lot of errors can be recovered from (eg, skipping over broken calls until we hit more working ones)
				steppingErr <- runErr
			}
			processedCount.Add(1)
		}(current)
	}

	return <-steppingErr
}

// will keep waiting until it reaches an error, fills a block, or is told to stop
func (csh *ContractStateHolder) RunOnUntil(state *statedb.StateDB, transactionCount int, reasonToQuit <-chan any) (transactions []utils.TransactionType, err error) {
	reasonToStop := make(chan (struct{}))
	successfulTransactions := safe.NewSafeSlice()
	csh.successfullyRunCount.Set(0)
	runCountChan := csh.successfullyRunCount.GetOnUpdate()
	steppingErr := make(chan error)
	go func() {
		//handle if we need to stop
		for {
			select {
			case count, ok := <-runCountChan:
				if !ok || count == transactionCount {
					reasonToStop <- struct{}{}
					return
				}
			case <-reasonToQuit:
				reasonToStop <- struct{}{}
				return
			case err = <-steppingErr:
				if err != nil {
					reasonToStop <- struct{}{}
					return
				}
			}
		}
	}()
	go func() {
		for _, current := range csh.contractsHeld {
			go func(ch *contractHeld) {
				//each held contract starts running on it's own until it has an error, or
				for currentPoint, endPoint := ch.getCurrentAndMaxTransaction(); currentPoint < endPoint; currentPoint, endPoint = ch.getCurrentAndMaxTransaction() {
					select {
					case <-reasonToStop:
						return
					default:
						if err := ch.Step(state, successfulTransactions); err != nil {
							//TODO: we can skip a lot of the more expected errors. EG, if you run out of funds, that should just be dropped. not stop everyone after
							steppingErr <- err
							return
						}
					}
				}
			}(current)
		}
	}()
	<-reasonToStop
	csh.lock.Lock()
	defer csh.lock.Unlock()
	//we only lock here so that any transactions added later on can still be added.
	transactions = make([]utils.TransactionType, successfulTransactions.Len())
	successfulTransactions.ForEach(func(i int, a any) bool {
		//reformat everything from any, to something
		transactions[i] = a.(utils.TransactionType)
		return true
	})
	return transactions, err
}
