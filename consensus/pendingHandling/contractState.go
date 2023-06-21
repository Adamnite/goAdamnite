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
		processStep:    make(chan processSteps),
		isPlaceholder:  false,
	}
	go func() { ct.processStep <- waitingToStart }()
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
	lock         sync.Mutex            //only one thread is to touch this at a time!
	vm           *VM.Machine           //the vm with all of these changes
	transactions []*complexTransaction //the transactions entered
	nextToRun    int                   //index of the next transaction to process
	runningHash  []byte
}

func newContractHeld(contract *common.Address, db VM.DBInterfaceItem) (*contractHeld, error) {
	vm, err := VM.NewVirtualMachineWithSpoofedConnection(contract, db)
	if err != nil {
		return nil, err
	}
	ch := contractHeld{
		vm:           vm,
		transactions: []*complexTransaction{},
		nextToRun:    0,
		runningHash:  vm.GetContractHash().Bytes(),
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
func (ch *contractHeld) Step(state *statedb.StateDB) error {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	if ch.nextToRun >= len(ch.transactions) {
		return fmt.Errorf("out of transactions to run. Current next up %d of %d", ch.nextToRun, len(ch.transactions))
	}
	hash, err := ch.transactions[ch.nextToRun].RunOn(ch.vm)
	ch.nextToRun += 1
	if err != nil {
		return err
	}
	ch.runningHash = crypto.Sha512(append(ch.runningHash, hash...))
	return nil
}

func (ch *contractHeld) RunAll(state *statedb.StateDB) error {
	//do this weirdness so that we can get the current point and the endpoint, while locking up the contract held as little as possible
	for currentPoint, endPoint := ch.getCurrentAndMaxTransaction(); currentPoint < endPoint; currentPoint, endPoint = ch.getCurrentAndMaxTransaction() {
		if err := ch.Step(state); err != nil {
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
	lock          sync.Mutex
	contractsHeld map[string]*contractHeld         //contractAddressString(hex)->contractHeld
	sends         map[string][]*complexTransaction //from(hexOfAddress)->all instances they have sent to, in order
	receives      map[string][]*complexTransaction //from(hexOfAddress)->all instances in which  they receive nite
	dbCache       VM.DBCacheAble                   //for keeping a local reference. allows each call to be made once
}

func NewContractStateHolder(dbEndpoint string) (*ContractStateHolder, error) {
	cache := VM.NewDBCache(dbEndpoint)
	csh := ContractStateHolder{
		contractsHeld: map[string]*contractHeld{},
		sends:         map[string][]*complexTransaction{},
		receives:      map[string][]*complexTransaction{},
		dbCache:       cache,
	}
	return &csh, nil
}
func (csh *ContractStateHolder) QueueTransaction(t *utils.VMCallTransaction) (err error) {
	//assume that the transaction is indeed, intended for us to handle it
	//also assume that the transactions are being fed to us in order
	csh.lock.Lock()
	defer csh.lock.Unlock()
	//check if the target contract is already loaded
	vmT := t.VMInteractions
	ch, exists := csh.contractsHeld[vmT.ContractCalled.Hex()]
	if !exists {
		//it's not stored yet, so we need to make this
		ch, err = newContractHeld(&vmT.ContractCalled, csh.dbCache)
		if err != nil {
			return err
		}
		csh.contractsHeld[vmT.ContractCalled.Hex()] = ch
	}

	//then check if the target method is loaded
	if err := csh.dbCache.PreCacheCode(vmT.ParametersPassed[:utils.FunctionIdentifierLength]); err != nil {
		return err
	}
	//TODO: check if the method calls another contract. If it does, we need to add a special point in that contracts queue
	var sendsMoney bool = t.Amount.Cmp(big.NewInt(0)) != 0
	var sendingBefore *complexTransaction = nil
	//then we actually log the transaction (so we know it'll work)
	if sendsMoney {
		//they send money... so we need to check if they're sending before. Or if they're receiving before
		from := t.From.GetAddress().Hex()
		if previouslySent, exists := csh.sends[from]; exists {
			sendingBefore = previouslySent[len(previouslySent)-1]
		} else if previouslyReceived, exists := csh.receives[from]; exists {
			sendingBefore = previouslyReceived[len(previouslyReceived)-1]
		}
	}
	ct, err := newComplexTransaction(t, sendingBefore)
	if err != nil {
		return err
	}
	ch.AddTransactionToQueue(ct)
	if sendsMoney {
		//we have to add that they sent nite, with reference to when they do, so it can be added.
		from := t.From.GetAddress().Hex()
		if previouslySent, exists := csh.sends[from]; exists {
			csh.sends[from] = append(previouslySent, ct)
		} else {
			csh.sends[from] = []*complexTransaction{ct}
		}

		to := t.To.Hex()
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
	processedCount := 0
	go func() {
		//this thread is setup to handle if we complete everything
		for processedCount < len(csh.contractsHeld) {
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
			processedCount++

		}(current)

	}

	return <-steppingErr
}
