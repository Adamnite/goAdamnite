package consensus

import (
	"bytes"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/utils/bytes"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/safe"
	"golang.org/x/sync/syncmap"
)

type round_data struct {
	lock              sync.RWMutex
	eligibleWitnesses *safe.SafeSlice //TODO: change these back to arrays and handle the rounds locking here
	witnesses         *safe.SafeSlice
	leadWitnessOrder  *safe.SafeSlice
	blocksPerWitness  []safe.SafeInt
	currentLeadIndex  safe.SafeInt
	blocksThisLead    safe.SafeInt
	witnessesMap      syncmap.Map //map[witnessPub]->*witness
	removedWitnesses  syncmap.Map //map[witnessPub]->the block number that they were removed in
	votes             syncmap.Map //map[witnessPub]->votes for that witness, in that round
	valueTotals       syncmap.Map //map[witnessPub] -> total amount staked on them
	vrfValues         syncmap.Map //map[witnessPub]->vrf value
	vrfProofs         syncmap.Map //map[witnessPub]->vrf proof
	vrfCutoffs        syncmap.Map //map[witnessPub]->vrf cutoff value
	seed              []byte
	openToApply       bool
	roundStartTime    time.Time
	blocksInRound     safe.SafeInt
}

func newRoundData(seed []byte) *round_data {
	newRound := round_data{
		lock:              sync.RWMutex{},
		eligibleWitnesses: safe.NewSafeSlice(),
		witnesses:         safe.NewSafeSlice(),
		leadWitnessOrder:  safe.NewSafeSlice(),
		witnessesMap:      syncmap.Map{},
		removedWitnesses:  syncmap.Map{},
		votes:             syncmap.Map{},
		valueTotals:       syncmap.Map{},
		vrfValues:         syncmap.Map{},
		vrfProofs:         syncmap.Map{},
		vrfCutoffs:        syncmap.Map{},
		openToApply:       true,
		seed:              seed,
		roundStartTime:    time.Now().UTC().Add(maxTimePrecision.Duration()).Truncate(maxTimePrecision.Duration()),
		//truncate the rounds start time to the closest reliable precision we can use.
		blocksInRound: *safe.NewSafeInt(0),
	}
	return &newRound
}
func (rd *round_data) getNextRoundSeed() []byte {
	rd.lock.RLock()
	defer rd.lock.RUnlock()
	// wit := rd.witnesses.Get(0).(*witness)
	// val, exists := rd.vrfValues.Load(wit.spendingPubString())
	// if !exists {
	// 	return []byte{}
	// }

	// return crypto.Sha512(val.([]byte))
	return crypto.Sha512(rd.seed)
}
func (rd *round_data) SetSeed(seed []byte) {
	rd.lock.Lock()
	defer rd.lock.Unlock()
	rd.seed = seed
}
func (rd *round_data) GetSeed() []byte {
	rd.lock.RLock()
	defer rd.lock.RUnlock()
	return rd.seed
}
func (rd *round_data) GetStartTime() time.Time {
	rd.lock.RLock()
	defer rd.lock.RUnlock()
	return rd.roundStartTime
}
func (rd *round_data) SetStartTime(rst time.Time) {
	rd.lock.Lock()
	defer rd.lock.Unlock()
	rd.roundStartTime = rst
}
func (rd *round_data) addEligibleWitness(w *witness, vrfVal []byte, vrfProof []byte) {
	rd.lock.Lock()
	defer rd.lock.Unlock()
	//since every candidate needs a vote (from themselves) to run, we can check the vote count
	_, exists := rd.votes.Load(w.spendingPubString())
	if exists {
		//witness already has votes, so they must exist
		log.Println("witness already has votes, so must already exist")
		return
	}
	rd.eligibleWitnesses.Append(w)
	rd.vrfValues.Store(w.spendingPubString(), vrfVal)
	rd.vrfProofs.Store(w.spendingPubString(), vrfProof)
	rd.valueTotals.Store(w.spendingPubString(), big.NewInt(0))
}
func (rd *round_data) addVote(v *utils.Voter) {
	rd.lock.Lock()
	defer rd.lock.Unlock()
	candidateId := (*crypto.PublicKey)(&v.To)
	votes, exists := rd.votes.Load(string(*candidateId))
	if !exists {
		votes = []*utils.Voter{}
	}
	votes = append(votes.([]*utils.Voter), v)
	rd.votes.Store(string(*candidateId), votes)
	valueTotal, exists := rd.valueTotals.Load(string(*candidateId))
	if !exists {
		valueTotal = big.NewInt(0)
	}
	valueTotal = big.NewInt(0).Add(valueTotal.(*big.Int), v.StakingAmount)
	rd.valueTotals.Store(string(*candidateId), valueTotal)
}
func (rd *round_data) BlockReviewed() {
	rd.lock.Lock()
	defer rd.lock.Unlock()

	rd.blocksInRound.Add(1)
	rd.blocksThisLead.Add(1)
	if rd.blocksPerWitness[rd.currentLeadIndex.Get()].Get() <= rd.blocksThisLead.Get() {
		rd.blocksThisLead.Set(0)
		rd.currentLeadIndex.Set((rd.currentLeadIndex.Get() + 1) % rd.leadWitnessOrder.Len())
	}
}
func (rd *round_data) GetVotesFor(w crypto.PublicKey) []*utils.Voter {
	rd.lock.RLock()
	defer rd.lock.RUnlock()
	votes, exists := rd.votes.Load(string(w))
	if !exists {
		return nil
	}
	return votes.([]*utils.Voter)
}

// select witnesses should be called only once all votes have been received, and the next round needs to start. Returns witnesses selected
func (rd *round_data) selectWitnesses(goalCount int) (*safe.SafeSlice, []byte) {
	if rd.openToApply { //close the applications and select the winning witnesses
		rd.openToApply = false
		maxBlockValidationPercent, maxStaking, maxVotes, maxElected := rd.getMaxes()
		rd.eligibleWitnesses.ForEach(func(_ int, val any) bool {
			w := val.(*witness)
			weight := rd.getWeight(w, maxBlockValidationPercent, maxStaking, maxVotes, maxElected)
			rd.vrfCutoffs.Store(w.spendingPubString(), weight)
			return true
		})
	} else {
		//no need to do the work again
		return rd.witnesses, rd.getNextRoundSeed()
	}
	if rd.eligibleWitnesses.Len() == 0 {
		//trying to prevent a null pointer error from happening if no one applied
		return nil, nil
	}
	rd.lock.Lock()

	passingWitnesses := safe.NewSafeSlice()
	witnessVrfValueFloat := make(map[string]*big.Float)
	//32 bytes, all set to 255
	maxVRFVal := big.NewFloat(0).SetInt(big.NewInt(1).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}))
	rd.eligibleWitnesses.ForEach(func(_ int, val any) bool {
		w := val.(*witness)
		vrfVal, _ := rd.vrfValues.Load(w.spendingPubString())
		witnessVRFValue := big.NewFloat(0).Quo(

			big.NewFloat(1).SetInt(big.NewInt(1).SetBytes(vrfVal.([]byte))),
			maxVRFVal,
		)
		witnessVrfValueFloat[string(w.spendingPub)] = witnessVRFValue
		vrfCut, _ := rd.vrfCutoffs.Load(w.spendingPubString())
		var vrfCutFloat *big.Float = vrfCut.(*big.Float)
		if vrfCutFloat.Cmp(witnessVRFValue) != -1 { //the witnesses VRF value is less than or equal to their cutoff
			passingWitnesses.Append(w)
		}
		return true
	})
	if rd.eligibleWitnesses.Len() <= goalCount { //we don't have enough witnesses to try and generate more
		passingWitnesses = rd.eligibleWitnesses.Copy()
	}
	//sort the passing witnesses based on the difference between the cutoff and their score
	passingWitnesses.Sort(func(a any, b any) bool {
		//we sort no matter what so that we can get the next rounds seed
		aId := a.(*witness)
		bId := b.(*witness)
		aOff, _ := rd.vrfCutoffs.Load(aId.spendingPubString())
		bOff, _ := rd.vrfCutoffs.Load(bId.spendingPubString())
		aDif := big.NewFloat(0).Sub(aOff.(*big.Float), witnessVrfValueFloat[aId.spendingPubString()])
		bDif := big.NewFloat(0).Sub(bOff.(*big.Float), witnessVrfValueFloat[bId.spendingPubString()])
		return aDif.Cmp(bDif) == 1
	})

	if passingWitnesses.Len() > goalCount && rd.eligibleWitnesses.Len() >= goalCount {
		//remove the lowest scorers
		passingWitnesses.RemoveFrom(goalCount, -1)
	} else if passingWitnesses.Len() < goalCount && rd.eligibleWitnesses.Len() >= goalCount {
		// get more witnesses
		rd.eligibleWitnesses.Sort(func(a any, b any) bool {
			//sort the eligible witnesses so that the ones closest to passing are at the beginning
			aId := a.(*witness)
			bId := b.(*witness)
			aOff, _ := rd.vrfCutoffs.Load(aId.spendingPubString())
			bOff, _ := rd.vrfCutoffs.Load(bId.spendingPubString())
			aDif := big.NewFloat(0).Sub(aOff.(*big.Float), witnessVrfValueFloat[aId.spendingPubString()])
			bDif := big.NewFloat(0).Sub(bOff.(*big.Float), witnessVrfValueFloat[bId.spendingPubString()])
			return aDif.Cmp(bDif) == 1
		})
		for i := 0; passingWitnesses.Len() < goalCount; i++ {
			passingWitnesses.Append(rd.eligibleWitnesses.Get(i))
		}
	}

	rd.witnesses = passingWitnesses
	rd.leadWitnessOrder = passingWitnesses.Copy()
	//create and copy the witnesses, then sorts it by their vrf values
	rd.leadWitnessOrder.Sort(func(aVal any, bVal any) bool {
		a := aVal.(*witness)
		b := bVal.(*witness)
		aVRFVal, _ := rd.vrfValues.Load(a.spendingPubString())
		bVRFVal, _ := rd.vrfValues.Load(b.spendingPubString())
		return bytes.Compare(aVRFVal.([]byte), bVRFVal.([]byte)) == -1

	})
	perWitness := maxBlocksPerRound / uint64(rd.eligibleWitnesses.Len())
	roundingCutoff := perWitness + (maxBlocksPerRound % uint64(rd.eligibleWitnesses.Len()))
	rd.blocksPerWitness = make([]safe.SafeInt, rd.leadWitnessOrder.Len())
	rd.leadWitnessOrder.ForEach(func(i int, val any) bool {
		if i == 0 {
			rd.blocksPerWitness[0].Set(int(roundingCutoff))
		} else {
			rd.blocksPerWitness[i].Set(int(perWitness))
		}
		return true
	})
	passingWitnesses.ForEach(func(i int, val any) bool {
		w := val.(*witness)
		rd.witnessesMap.Store(string(w.spendingPub), w)
		return true
	})
	rd.lock.Unlock()
	return passingWitnesses, rd.getNextRoundSeed()
}

func (rd *round_data) RemoveSelectedWitness(wit *witness, blockID uint64) error {
	rd.lock.Lock()
	defer rd.lock.Unlock()
	//TODO: check that the witness is able to be removed from this round without error
	rd.removedWitnesses.Store(wit.spendingPubString(), blockID)
	rd.witnessesMap.Delete(string(wit.spendingPub))
	//assume they can be removed safely

	//now we need to change the upcoming order for the witnesses.
	currentLead := rd.leadWitnessOrder.Get(rd.currentLeadIndex.Get()).(*witness)
	if bytes.Equal(currentLead.spendingPub, wit.spendingPub) {
		//they're actively the one running
		extraBlocks := uint64(rd.blocksPerWitness[rd.currentLeadIndex.Get()].Get()) - uint64(rd.blocksThisLead.Get())
		rd.blocksPerWitness[rd.currentLeadIndex.Get()].Set(rd.blocksThisLead.Get()) // stop them at whatever block they're at now
		//if they're last, the first person gets all of the blocks, otherwise everyone after them gets them
		if rd.currentLeadIndex.Get() == rd.leadWitnessOrder.Len()-1 {
			//set the first lead back to where they were, then have them pick up where they were.
			rd.currentLeadIndex.Set(0)
			rd.blocksThisLead.Set(rd.blocksPerWitness[0].Get())
			rd.blocksPerWitness[0].Add(int(extraBlocks))
			return nil
		}
		//assume they weren't the last, so we add the blocks in order to everyone next
		rd.currentLeadIndex.Add(1)
		remaining := rd.leadWitnessOrder.Len() - rd.currentLeadIndex.Get()
		for i := 0; i < int(extraBlocks); i++ {
			rd.blocksPerWitness[(i%remaining)+rd.currentLeadIndex.Get()].Add(1)
		}
		return nil
	}
	//TODO: test, and figure out this edge case
	//we know how many blocks we have left
	return nil
}
func (rd *round_data) IsActiveWitnessLead(witID *crypto.PublicKey) bool {
	rd.lock.RLock()
	defer rd.lock.RUnlock()
	if rd.openToApply {
		//if you can still apply, then you obviously cant have a lead for this round yet
		return false
	}
	// log.Printf("if this is above an error, the round is %v", rd)
	currentLead := rd.leadWitnessOrder.Get(rd.currentLeadIndex.Get())
	if currentLead == nil {
		return false
	}
	return bytes.Equal(currentLead.(*witness).spendingPub, *witID)
}

// func (rd *round_data) WasWitnessLead(witID *crypto.PublicKey, blockNumber *big.Int) bool {
// 	if rd.openToApply {
// 		//if you can still apply, then you obviously couldn't have been lead
// 		return false
// 	}
// 	rd.
// }

func (rd *round_data) getWeight(w *witness, maxBlockValidationPercent float64, maxStaking *big.Int, maxVotes int, maxElected uint64) *big.Float {
	rd.lock.RLock()
	defer rd.lock.RUnlock()
	tot, _ := rd.valueTotals.Load(w.spendingPubString())
	avgStakingAmount := float64(math.GetPercent(tot.(*big.Int), maxStaking))
	avgBlockValidationPercent := float64(w.validationPercent()) / float64(maxBlockValidationPercent)
	votes, _ := rd.votes.Load(w.spendingPubString())
	avgVoterCount := float64(len(votes.([]*utils.Voter))) / float64(maxVotes)
	var avgElectedCount float64
	if maxElected == 0 {
		avgElectedCount = 0
	} else {
		avgElectedCount = float64(w.timesElected) / float64(maxElected)
	}

	return utils.VRFCutoff(avgStakingAmount, avgBlockValidationPercent, avgVoterCount, avgElectedCount)

}

// find the largest values from a list of witnesses.
func (rd *round_data) getMaxes() (maxBlockValidationPercent float64, maxStakingAmount *big.Int, maxVoterCount int, maxElectedCount uint64) {
	rd.lock.RLock()
	defer rd.lock.RUnlock()
	maxStakingAmount = big.NewInt(0)
	maxBlockValidationPercent = 0.0
	maxVoterCount = 0
	maxElectedCount = 0
	rd.eligibleWitnesses.ForEach(func(i int, val any) bool {
		w := val.(*witness)
		if t := w.validationPercent(); maxBlockValidationPercent < t {
			maxBlockValidationPercent = t
		}
		t, _ := rd.valueTotals.Load(w.spendingPubString())
		if maxStakingAmount.Cmp(t.(*big.Int)) == -1 {
			maxStakingAmount = t.(*big.Int)
		}
		v, _ := rd.votes.Load(w.spendingPubString())
		if maxVoterCount < len(v.([]*utils.Voter)) {
			maxVoterCount = len(v.([]*utils.Voter))
		}

		if maxElectedCount < w.timesElected {
			maxElectedCount = w.timesElected
		}
		return true
	})
	return
}
