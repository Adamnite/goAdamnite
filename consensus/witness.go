package consensus

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/safe"
	"golang.org/x/sync/syncmap"
)

type witness struct {
	vrfKey crypto.PublicKey //our VRF public Key

	spendingPub crypto.PublicKey

	blocksReviewed uint64
	blocksApproved uint64
	timesElected   uint64
}

func (w witness) spendingPubString() string {
	return string(w.spendingPub)
}

func (w witness) validationPercent() float64 {
	if w.blocksReviewed == 0 {
		return 1
	}
	return float64(w.blocksApproved) / float64(w.blocksReviewed)
}
func witnessFromCandidate(can *utils.Candidate) *witness {
	w := witness{
		vrfKey:      can.VRFKey,
		spendingPub: *can.GetWitnessPub(),

		blocksReviewed: 0,
		blocksApproved: 0,
		timesElected:   0,
	}

	return &w
}

type Witness_pool struct {
	witnessGoal           int
	totalCandidates       map[string]*utils.Candidate //witID->Candidate. Use for verifying votes
	totalWitnesses        map[string]*witness         //witID-> witness
	rounds                syncmap.Map                 //round ID ->data
	currentWorkingRoundID *safe.SafeInt               //the round that is currently working. Next round should be accepting
	consensusType         uint8                       //support type that this is being pitched for

	newRoundStartedCaller []func()
	asyncTrackingRunning  bool
	asyncStopper          context.CancelFunc
	lock                  sync.RWMutex //im tired of everything having race conditions
}

func NewWitnessPool(roundNumber int, consensusType networking.NetworkTopLayerType, seed []byte) (*Witness_pool, error) {
	wp := Witness_pool{
		witnessGoal:           27,
		totalCandidates:       make(map[string]*utils.Candidate),
		totalWitnesses:        make(map[string]*witness),
		rounds:                syncmap.Map{},
		consensusType:         uint8(consensusType),
		currentWorkingRoundID: safe.NewSafeInt(roundNumber),
		asyncTrackingRunning:  false,
		lock:                  sync.RWMutex{},
	}
	wp.newRound(roundNumber, seed)
	wp.newRound(roundNumber+1, seed)
	// if err := wp.newRound(roundNumber, seed); err != nil {
	// 	return nil, err
	// }

	return &wp, wp.StartAsyncTracking()
}
func (wp *Witness_pool) CatchupTo(oldestUsefulRound uint64, block *utils.Block, chain *blockchain.Blockchain) error {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	if chain == nil {
		return errors.New("chain not set")
	}
	tmp := block
	for (tmp.Header.ParentBlockID != common.Hash{}) { //the Genesis Block hash
		//TODO: figure out what round these blocks belong to
		//TODO: add these witnesses approval rating to their witness stats
		//TODO: create an in memory logging of their actions
		parentBlock := chain.GetBlockByHash(tmp.Header.ParentBlockID)
		if parentBlock == nil {
			// parent block does not exist on chain
			return errors.New("parent block was not found")
		}
		// note: temporary adapter until we start using consensus structures across the rest of codebase
		tmp = ConvertBlock(parentBlock)
	}
	return nil
}

// starts a new thread to keep track of the round, primarily that it does not go over time
func (wp *Witness_pool) StartAsyncTracking() error {
	//TODO: probably need to lock this
	if wp.asyncTrackingRunning {
		return fmt.Errorf("already has an async tracker going")
	}
	ctx, canceler := context.WithCancel(context.Background())

	wp.asyncStopper = canceler
	wp.asyncTrackingRunning = true
	//run our continuos loop that checks every time the last round should've stopped.
	go func(ctx context.Context) {
		for {
			roundEndTime := wp.GetWorkingRound().GetStartTime().Add(maxTimePerRound.Duration()).Truncate(maxTimePrecision.Duration())
			roundThatStartedThis := wp.currentWorkingRoundID
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Until(roundEndTime)):
				//wait until the max end point of the last round
				//double check it waited the correct time (and a new round wasn't started without us noticing)
				if time.Now().After(roundEndTime) && wp.asyncTrackingRunning && wp.currentWorkingRoundID == roundThatStartedThis {
					//there's a lot of time between rounds. something could've canceled this without us noticing
					//the time has actually elapsed. Meaning that round took the max time
					wp.nextRound()
				}
			}
		}
	}(ctx)
	return nil
}
func (wp *Witness_pool) StopAsyncTracker() {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	wp.asyncTrackingRunning = false
	if wp.asyncStopper != nil {
		wp.asyncStopper()
		wp.asyncStopper = nil
	}

}

// call after a witness reviews a block. Call once per block
func (wp *Witness_pool) ActiveWitnessReviewed(witID *crypto.PublicKey, successful bool, blockID uint64) error {
	wp.lock.RLock()
	wit := wp.totalWitnesses[string(*witID)]
	wp.lock.RUnlock()
	if wit == nil {
		return errors.New("witness is not stored locally") //TODO:change to real error
	}
	if !wp.IsActiveWitness(witID) {
		return fmt.Errorf("witness is not running in reviewed round")
	}
	if !wp.IsActiveWitnessLead(witID) {
		return fmt.Errorf("witness is not lead presently")
	}

	//by here, we know the witness is running in the round they say they are.
	currentWorkingRound := wp.GetWorkingRound()
	wit.blocksReviewed += 1
	if successful {
		wit.blocksApproved += 1
	}
	if !successful {
		//the block seems to be faulty. Assume this witness misbehaved, remove them, and ignore the block
		return wp.UpdateWorkingRound(func(rd *round_data) error {
			return rd.RemoveSelectedWitness(wit, blockID)
		})
	}

	wp.UpdateWorkingRound(func(rd *round_data) error {
		rd.BlockReviewed()
		return nil
	})
	if uint64(currentWorkingRound.blocksInRound.Get()) >= maxBlocksPerRound {
		//the round has reached its block limit.
		wp.nextRound()
	}
	return nil
}

// gets the round. If the round isn't already stored, we make one :)
func (wp *Witness_pool) getRound(roundID int) *round_data {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	if roundID < 0 {
		return nil
	}
	rd, exists := wp.rounds.Load(roundID)
	if exists {
		return rd.(*round_data)
	}
	//the round isn't stored locally, but i believe it's easiest to just create it then and store it
	newRD := newRoundData([]byte{}) //we cant know the seed, as such, the seed should be instead updated each round
	newRD.openToApply = roundID > wp.currentWorkingRoundID.Get()
	//just make sure that loading old rounds cant receive more votes or anything
	wp.rounds.Store(roundID, newRD)
	return newRD
}

// locks the thread and handles updating the round.
func (wp *Witness_pool) updateRound(roundID int, updating func(*round_data) error) error {
	rd := wp.getRound(roundID)
	wp.lock.Lock()
	defer wp.lock.Unlock()
	err := updating(rd)
	if err != nil {
		return err
	}
	wp.rounds.Store(roundID, rd)
	return nil
}

// gets the current round accepting votes
func (wp *Witness_pool) GetApplyingRound() *round_data {
	return wp.getRound(wp.currentWorkingRoundID.Get() + 1)
}

// update the current applying round
func (wp *Witness_pool) UpdateApplyingRound(updating func(*round_data) error) error {
	return wp.updateRound(wp.currentWorkingRoundID.Get()+1, updating)
}

// gets the working round. AKA, the one with the active witnesses
func (wp *Witness_pool) GetWorkingRound() *round_data {
	return wp.getRound(wp.currentWorkingRoundID.Get())
}
func (wp *Witness_pool) UpdateWorkingRound(updating func(*round_data) error) error {
	return wp.updateRound(wp.currentWorkingRoundID.Get(), updating)
}
func (wp *Witness_pool) newRound(roundID int, seed []byte) error {
	//this way, if that round already existed, instead we are updating it
	return wp.updateRound(roundID, func(rd *round_data) error {
		rd.SetSeed(seed)
		return nil
	})
}

// stats the next round.
func (wp *Witness_pool) nextRound() {
	var nextSeed []byte
	if wp.currentWorkingRoundID.Get() == 0 {
		nextSeed = []byte{} //seeds from the initial round is 0
		wp.UpdateApplyingRound(func(rd *round_data) error {
			rd.openToApply = false
			return nil
		})
	} else {
		_, nextSeed = wp.SelectCurrentWitnesses()
	}

	wp.currentWorkingRoundID.Add(1)
	if err := wp.newRound(wp.currentWorkingRoundID.Get()+1, nextSeed); err != nil {
		//add a new applying round
		log.Println(err)
	}

	wp.UpdateWorkingRound(func(rd *round_data) error {
		rd.SetStartTime(time.Now().UTC().Truncate(maxTimePrecision.Duration()))
		return nil
	})

	for _, nextRoundFunc := range wp.newRoundStartedCaller {
		nextRoundFunc()
	}
}
func (wp *Witness_pool) AddNewRoundCaller(f func()) {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	wp.newRoundStartedCaller = append(wp.newRoundStartedCaller, f)
}
func (wp *Witness_pool) GetCandidate(witID *crypto.PublicKey) *utils.Candidate {
	wp.lock.RLock()
	defer wp.lock.RUnlock()
	return wp.totalCandidates[string(*witID)]
}

// get the most recent seed needed to apply.
func (wp *Witness_pool) GetCurrentSeed() []byte {
	rd := wp.GetApplyingRound()
	if rd == nil {
		return []byte{} //
	}

	if rd.openToApply {
		return rd.GetSeed()
	} else {
		//the current round must be in error, as it's already closed. Start the next round
		log.Println("get current seed was called on a round that is not accepting candidates")
		wp.nextRound() //TODO: this seems like it's a bad idea, right?
	}

	return wp.GetCurrentSeed()
}

func (wp *Witness_pool) AddCandidate(can *utils.Candidate) error {
	wp.lock.Lock()
	wit := wp.totalWitnesses[can.GetWitnessString()]
	if wit == nil || !bytes.Equal(wit.spendingPub, *can.GetWitnessPub()) {
		//new candidate
		wit = witnessFromCandidate(can)
		wp.totalWitnesses[can.GetWitnessString()] = wit
		wp.totalCandidates[can.GetWitnessString()] = can
	}
	//by only creating the witness if it's a new witID, we prevent changing of VRFKeys every round to get the best outcome
	wp.lock.Unlock()
	canRound := wp.getRound(int(can.Round))

	if !canRound.openToApply {
		//this is most likely just an old candidate, or someone trying to pitch for far out
		log.Printf("unable to apply for that round, the candidate is proposing for round %v, and the current working roundID is %v", can.Round, wp.currentWorkingRoundID.Get())
		return fmt.Errorf("candidates application for this round does not make sense") //TODO: change to real error
	} else {
		//we also verify their VRF
		if !wit.vrfKey.Verify(canRound.GetSeed(), can.VRFValue, can.VRFProof) {
			//they lied about their VRFValue
			return fmt.Errorf("candidate's VRF is unverifiable") //TODO: change to real error
		}
	}
	wp.updateRound(int(can.Round), func(rd *round_data) error {
		rd.addEligibleWitness(wit, can.VRFValue, can.VRFProof)
		rd.addVote(&can.InitialVote)
		return nil
	})
	return nil
}
func (wp *Witness_pool) AddVoteForCurrent(vote *utils.Voter) error {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	wp.UpdateApplyingRound(func(rd *round_data) error {
		rd.addVote(vote)
		return nil
	})

	return nil //TODO: return an error if the rounds system is broken
}
func (wp *Witness_pool) AddVote(round uint64, v *utils.Voter) error {
	return wp.updateRound(int(round), func(rd *round_data) error {
		if !rd.openToApply {
			return fmt.Errorf("round selected is not accepting votes")
		}
		rd.addVote(v)
		return nil
	})
}

// select closes the current rounds submissions, and starts the submissions for the next round
func (wp *Witness_pool) SelectCurrentWitnesses() (witnesses *safe.SafeSlice, newSeed []byte) {
	wp.UpdateApplyingRound(func(rd *round_data) error {
		witnesses, newSeed = rd.selectWitnesses(wp.witnessGoal)
		return nil
	})
	return
}

// returns if this is not only an active witness this round, but if they are the lead witness
func (wp *Witness_pool) IsActiveWitnessLead(witID *crypto.PublicKey) bool {
	if !wp.IsActiveWitness(witID) {
		//if they aren't active rn, then we know they they cant be the lead
		return false
	}
	return wp.GetWorkingRound().IsActiveWitnessLead(witID)
}

// func (wp *Witness_pool) WasWitnessLead(witID *crypto.PublicKey, roundNumber int, blockNumber *big.Int) bool {
// 	rd := wp.getRound(roundNumber)
// }

// returns if the witness is in the calculating round (note, this does not show if the witness has applied to run again next round)
func (wp *Witness_pool) IsActiveWitness(witID *crypto.PublicKey) bool {
	workingRound := wp.GetWorkingRound()
	wp.lock.RLock()
	defer wp.lock.RUnlock()
	if wp.totalCandidates[string(*witID)] == nil {
		// might as well eliminate witnesses we don't have at all as quickly as possible
		return false
	}
	_, exists := workingRound.witnessesMap.Load(string(*witID))
	return exists
}
