package consensus

import (
	"errors"
	"fmt"
	"time"

	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
)

type witness struct {
	vrfKey crypto.PublicKey //our VRF public Key

	networkString string
	nodeID        crypto.PublicKey

	blocksReviewed uint64
	blocksApproved uint64
	timesElected   uint64
}

func (w witness) validationPercent() float64 {
	if w.blocksReviewed == 0 {
		return 1
	}
	return float64(w.blocksApproved) / float64(w.blocksReviewed)
}
func witnessFromCandidate(can *utils.Candidate) *witness {
	w := witness{
		vrfKey:        can.VRFKey,
		networkString: can.NetworkString,
		nodeID:        can.NodeID,

		blocksReviewed: 0,
		blocksApproved: 0,
		timesElected:   0,
	}

	return &w
}

type Witness_pool struct {
	witnessGoal     int
	totalCandidates map[string]*utils.Candidate //nodeID->Candidate. Use for verifying votes
	totalWitnesses  map[string]*witness         //nodeID-> witness
	rounds          map[uint64]*round_data      //round ID ->data
	currentRound    uint64                      //the round that is currently accepting witness applications
	consensusType   uint8                       //support type that this is being pitched for

	newRoundStartedCaller func()
	asyncTrackingRunning  bool
}

func NewWitnessPool(roundNumber uint64, consensusType networking.NetworkTopLayerType, seed []byte) (*Witness_pool, error) {
	wp := Witness_pool{
		witnessGoal:          27,
		totalCandidates:      make(map[string]*utils.Candidate),
		totalWitnesses:       make(map[string]*witness),
		rounds:               map[uint64]*round_data{},
		consensusType:        uint8(consensusType),
		currentRound:         roundNumber,
		asyncTrackingRunning: false,
	}
	err := wp.newRound(roundNumber, seed)
	return &wp, err
}
func (wp *Witness_pool) CatchupTo(oldestUsefulRound uint64, block *utils.Block, chain *blockchain.Blockchain) error {
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
	if wp.asyncTrackingRunning {
		return fmt.Errorf("already has an async tracker going")
	}
	wp.asyncTrackingRunning = true
	//run our continuos loop that checks every time the last round should've stopped.
	go func() {
		for wp.asyncTrackingRunning {
			roundEndTime := wp.GetCurrentRound().roundStartTime.Add(maxTimePerRound)
			//wait until the max end point of the last round
			<-time.After(time.Until(roundEndTime))
			//double check it waited the correct time (and a new round wasn't started without us noticing)
			if time.Now().After(roundEndTime) && wp.asyncTrackingRunning { //there's a lot of time between rounds. something could've canceled this without us noticing
				//the time has actually elapsed. Meaning that round took the max time
				wp.SelectCurrentWitnesses()
			}
		}
	}()
	return nil
}
func (wp *Witness_pool) StopAsyncTracker() {
	wp.asyncTrackingRunning = false
}

// call after a witness reviews a block. Call once per block
func (wp *Witness_pool) ActiveWitnessReviewed(nodeID *crypto.PublicKey, successful bool) error {
	wit := wp.totalWitnesses[string(*nodeID)]
	if wit == nil {
		return errors.New("witness is not stored locally") //TODO:change to real error
	}
	if wp.IsActiveWitness(nodeID) {
		return fmt.Errorf("witness is not running in reviewed round")
	}

	//by here, we know the witness is running in the round they say they are.
	wit.blocksReviewed += 1
	if successful {
		wit.blocksApproved += 1
	}

	wp.GetCurrentRound().blocksInRound += 1
	return nil
}
func (wp *Witness_pool) GetCurrentRound() *round_data {
	return wp.rounds[wp.currentRound]
}
func (wp *Witness_pool) newRound(roundID uint64, seed []byte) error {
	if _, exists := wp.rounds[roundID]; exists {
		return fmt.Errorf("round already logged locally")
	}
	wp.rounds[roundID] = newRoundData(seed)
	return nil
}

// stats the next round.
func (wp *Witness_pool) nextRound() {
	wp.newRound(wp.currentRound+1, wp.GetCurrentRound().getNextRoundSeed())
	wp.currentRound += 1

	if wp.newRoundStartedCaller != nil {
		//TODO: call anyone who needs to know the round updated
		wp.newRoundStartedCaller()
	}
}
func (wp Witness_pool) GetCandidate(nodeID *crypto.PublicKey) *utils.Candidate {
	return wp.totalCandidates[string(*nodeID)]
}

// get the most recent seed needed to apply.
func (wp Witness_pool) GetCurrentSeed() []byte {
	rd := wp.GetCurrentRound()
	if rd == nil {
		return []byte{} //
	}
	if rd.openToApply {
		return rd.seed
	}

	wp.currentRound += 1
	return rd.getNextRoundSeed()

}

func (wp *Witness_pool) AddCandidate(can *utils.Candidate) error {
	wit := wp.totalWitnesses[string(can.NodeID)]
	if wit == nil {
		//new candidate
		wit = witnessFromCandidate(can)
		wp.totalWitnesses[string(can.NodeID)] = wit
		wp.totalCandidates[string(can.NodeID)] = can
	}
	//by only creating the witness if it's a new nodeID, we prevent changing of VRFKeys every round to get the best outcome

	if rd, exists := wp.rounds[can.Round]; !exists || !rd.openToApply {
		//this is most likely just an old candidate, or someone trying to pitch for far out
		return fmt.Errorf("candidates application for this round does not make sense") //TODO: change to real error
	} else {
		//we also verify their VRF
		if !wit.vrfKey.Verify(rd.seed, can.VRFValue, can.VRFProof) {
			//they lied about their VRFValue
			return fmt.Errorf("candidate's VRF is unverifiable") //TODO: change to real error
		}
	}

	wp.rounds[can.Round].addEligibleWitness(wit, can.VRFValue, can.VRFProof)
	wp.rounds[can.Round].addVote(&can.InitialVote)

	return nil
}
func (wp *Witness_pool) AddVoteForCurrent(vote *utils.Voter) error {
	wp.GetCurrentRound().addVote(vote)
	return nil //TODO: return an error if the rounds system is broken
}
func (wp *Witness_pool) AddVote(round uint64, v *utils.Voter) error {
	if wp.rounds[round] == nil {
		return fmt.Errorf("round selected is not recorded yet")
	}
	if !wp.rounds[round].openToApply {
		return fmt.Errorf("round selected is not accepting votes")
	}
	wp.rounds[round].addVote(v)
	return nil
}

// select closes the current rounds submissions, and starts the submissions for the next round
func (wp *Witness_pool) SelectCurrentWitnesses() (witnesses []*witness, newSeed []byte) {
	witnesses, newSeed = wp.GetCurrentRound().selectWitnesses(wp.witnessGoal)
	wp.nextRound()
	return
}

// returns if the witness is in the calculating round (note, this does not show if the witness has applied to run again next round)
func (wp Witness_pool) IsActiveWitness(nodeID *crypto.PublicKey) bool {
	return wp.WasWitnessActive(wp.currentRound-1, nodeID)
}
func (wp Witness_pool) WasWitnessActive(roundID uint64, nodeID *crypto.PublicKey) bool {
	if wp.totalCandidates[string(*nodeID)] == nil {
		// might as well eliminate witnesses we don't have at all as quickly as possible
		return false
	}

	rd, exists := wp.rounds[roundID]
	if !exists ||
		rd.openToApply {
		//round didn't exist, so no way they could've been running
		//if the round is open to applying, then there's no way that this round could've been verifying
		return false
	}
	//TODO: for checking witnesses that were removed mid round, we will also need the block number
	return rd.witnessesMap[string(*nodeID)] != nil
}
