package consensus

import (
	"bytes"
	"math/big"
	"sort"
	"time"

	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils"
)

type round_data struct {
	eligibleWitnesses []*witness
	witnesses         []*witness
	leadWitnessOrder  []*witness
	blocksPerWitness  []uint64
	currentLeadIndex  int
	blocksThisLead    uint64
	witnessesMap      map[string]*witness
	removedWitnesses  map[string]uint64         //map[witnessPub]->the block number that they were removed in
	votes             map[string][]*utils.Voter //map[witnessPub]->votes for that witness, in that round
	valueTotals       map[string]*big.Int       //map[witnessPub] -> total amount staked on them
	vrfValues         map[string][]byte         //map[witnessPub]->vrf value
	vrfProofs         map[string][]byte         //map[witnessPub]->vrf proof
	vrfCutoffs        map[string]*big.Float     //map[witnessPub]->vrf cutoff value
	seed              []byte
	openToApply       bool
	roundStartTime    time.Time
	blocksInRound     uint64
}

func newRoundData(seed []byte) *round_data {
	newRound := round_data{
		witnessesMap:     make(map[string]*witness),
		removedWitnesses: make(map[string]uint64),
		votes:            make(map[string][]*utils.Voter),
		valueTotals:      make(map[string]*big.Int),
		vrfValues:        make(map[string][]byte),
		vrfProofs:        make(map[string][]byte),
		vrfCutoffs:       make(map[string]*big.Float),
		openToApply:      true,
		seed:             seed,
		roundStartTime:   time.Now().UTC().Truncate(maxTimePrecision),
		//truncate the rounds start time to the closest reliable precision we can use.
		blocksInRound: 0,
	}
	return &newRound
}
func (rd *round_data) getNextRoundSeed() []byte {
	if len(rd.witnesses) == 0 {
		return []byte{}
	}
	return crypto.Sha512(rd.vrfValues[string(rd.witnesses[0].spendingPub)])
}
func (rd *round_data) addEligibleWitness(w *witness, vrfVal []byte, vrfProof []byte) {
	//since every candidate needs a vote (from themselves) to run, we can check the vote count
	if rd.votes[string(w.spendingPub)] != nil {
		return
	}

	rd.eligibleWitnesses = append(rd.eligibleWitnesses, w)
	rd.vrfValues[string(w.spendingPub)] = vrfVal
	rd.vrfProofs[string(w.spendingPub)] = vrfProof
	rd.valueTotals[string(w.spendingPub)] = big.NewInt(0)
}
func (rd *round_data) addVote(v *utils.Voter) {
	candidateId := (*crypto.PublicKey)(&v.To)
	if rd.votes[string(*candidateId)] == nil {
		rd.votes[string(*candidateId)] = []*utils.Voter{}
	}
	rd.votes[string(*candidateId)] = append(rd.votes[string(*candidateId)], v)
	if rd.valueTotals[string(*candidateId)] == nil {
		rd.valueTotals[string(*candidateId)] = big.NewInt(0)
	}
	rd.valueTotals[string(*candidateId)].Add(rd.valueTotals[string(*candidateId)], v.StakingAmount)
}
func (rd *round_data) BlockReviewed() {
	rd.blocksInRound += 1
	rd.blocksThisLead += 1
	if rd.blocksPerWitness[rd.currentLeadIndex] <= rd.blocksThisLead {
		rd.blocksThisLead = 0
		rd.currentLeadIndex = (rd.currentLeadIndex + 1) % len(rd.leadWitnessOrder)
	}
}

// select witnesses should be called only once all votes have been received, and the next round needs to start. Returns witnesses selected
func (rd *round_data) selectWitnesses(goalCount int) ([]*witness, []byte) {
	if rd.openToApply { //close the applications and select the winning witnesses
		rd.openToApply = false
		maxBlockValidationPercent, maxStaking, maxVotes, maxElected := rd.getMaxes()
		for _, w := range rd.eligibleWitnesses {
			weight := rd.getWeight(w, maxBlockValidationPercent, maxStaking, maxVotes, maxElected)
			rd.vrfCutoffs[string(w.spendingPub)] = weight
		}
	}
	// else {
	// 	//no need to do the work again
	// 	return rd.witnesses, crypto.Sha512(rd.vrfValues[string(rd.witnesses[0].spendingPub)])
	// }

	passingWitnesses := []*witness{}
	witnessVrfValueFloat := make(map[string]*big.Float)
	//32 bytes, all set to 255
	maxVRFVal := big.NewFloat(0).SetInt(big.NewInt(1).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}))
	for _, w := range rd.eligibleWitnesses {
		witnessVRFValue := big.NewFloat(0).Quo(
			big.NewFloat(1).SetInt(big.NewInt(1).SetBytes(rd.vrfValues[string(w.spendingPub)])),
			maxVRFVal,
		)
		witnessVrfValueFloat[string(w.spendingPub)] = witnessVRFValue
		if rd.vrfCutoffs[string(w.spendingPub)].Cmp(witnessVRFValue) != -1 { //the witnesses VRF value is less than or equal to their cutoff
			passingWitnesses = append(passingWitnesses, w)
		}
	}
	if len(rd.eligibleWitnesses) <= goalCount { //we don't have enough witnesses to try and generate more
		passingWitnesses = rd.eligibleWitnesses
	}
	//sort the passing witnesses based on the difference between the cutoff and their score
	sort.Slice(passingWitnesses, func(i, j int) bool {
		//we sort no matter what so that we can get the next rounds seed
		aId := passingWitnesses[i].spendingPub
		bId := passingWitnesses[j].spendingPub
		aDif := big.NewFloat(0).Sub(rd.vrfCutoffs[string(aId)], witnessVrfValueFloat[string(aId)])
		bDif := big.NewFloat(0).Sub(rd.vrfCutoffs[string(bId)], witnessVrfValueFloat[string(bId)])
		return aDif.Cmp(bDif) == 1
	})

	if len(passingWitnesses) > goalCount && len(rd.eligibleWitnesses) >= goalCount {
		//remove the lowest scorers
		passingWitnesses = passingWitnesses[:goalCount]
	} else if len(passingWitnesses) < goalCount && len(rd.eligibleWitnesses) >= goalCount {
		// get more witnesses
		sort.Slice(rd.eligibleWitnesses, func(i, j int) bool {
			//sort the eligible witnesses so that the ones closest to passing are at the beginning
			aId := rd.eligibleWitnesses[i].spendingPub
			bId := rd.eligibleWitnesses[j].spendingPub
			aDif := big.NewFloat(0).Sub(rd.vrfCutoffs[string(aId)], witnessVrfValueFloat[string(aId)])
			bDif := big.NewFloat(0).Sub(rd.vrfCutoffs[string(bId)], witnessVrfValueFloat[string(bId)])
			return aDif.Cmp(bDif) == 1
		})
		passingWitnesses = append(passingWitnesses,
			rd.eligibleWitnesses[len(passingWitnesses):goalCount]...)
	}

	rd.witnesses = passingWitnesses
	rd.leadWitnessOrder = make([]*witness, len(passingWitnesses))
	copy(rd.leadWitnessOrder, passingWitnesses)
	//create and copy the witnesses, then sorts it by their vrf values
	sort.Slice(rd.leadWitnessOrder, func(i, j int) bool {
		a := rd.leadWitnessOrder[i].spendingPub
		b := rd.leadWitnessOrder[j].spendingPub
		return bytes.Compare(rd.vrfValues[string(a)], rd.vrfValues[string(b)]) == -1
	})
	perWitness := maxBlocksPerRound / uint64(len(rd.eligibleWitnesses))
	roundingCutoff := perWitness + (maxBlocksPerRound % uint64(len(rd.eligibleWitnesses)))
	rd.blocksPerWitness = make([]uint64, len(rd.leadWitnessOrder))
	for i := range rd.leadWitnessOrder {
		if i == 0 {
			rd.blocksPerWitness[0] = roundingCutoff
		} else {
			rd.blocksPerWitness[i] = perWitness
		}
	}
	for _, w := range passingWitnesses {
		rd.witnessesMap[string(w.spendingPub)] = w
	}
	return passingWitnesses, crypto.Sha512(rd.vrfValues[string(passingWitnesses[0].spendingPub)])
}

func (rd *round_data) RemoveSelectedWitness(wit *witness, blockID uint64) error {
	//TODO: check that the witness is able to be removed from this round without error
	rd.removedWitnesses[string(wit.spendingPub)] = blockID
	delete(rd.witnessesMap, string(wit.spendingPub))
	//assume they can be removed safely

	//now we need to change the upcoming order for the witnesses.
	if bytes.Equal(rd.leadWitnessOrder[rd.currentLeadIndex].spendingPub, wit.spendingPub) {
		//they're actively the one running
		extraBlocks := rd.blocksPerWitness[rd.currentLeadIndex] - rd.blocksThisLead
		rd.blocksPerWitness[rd.currentLeadIndex] = rd.blocksThisLead // stop them at whatever block they're at now
		//if they're last, the first person gets all of the blocks, otherwise everyone after them gets them
		if rd.currentLeadIndex == len(rd.leadWitnessOrder)-1 {
			//set the first lead back to where they were, then have them pick up where they were.
			rd.currentLeadIndex = 0
			rd.blocksThisLead = rd.blocksPerWitness[0]
			rd.blocksPerWitness[0] += extraBlocks
			return nil
		}
		//assume they weren't the last, so we add the blocks in order to everyone next
		rd.currentLeadIndex += 1
		remaining := len(rd.leadWitnessOrder) - rd.currentLeadIndex
		for i := 0; i < int(extraBlocks); i++ {
			rd.blocksPerWitness[(i%remaining)+rd.currentLeadIndex] += 1
		}
		return nil
	}
	//TODO: test, and figure out this edge case
	//we know how many blocks we have left
	return nil
}
func (rd round_data) IsActiveWitnessLead(witID *crypto.PublicKey) bool {
	if rd.openToApply {
		//if you can still apply, then you obviously cant have a lead for this round yet
		return false
	}
	// log.Printf("if this is above an error, the round is %v", rd)
	return bytes.Equal(rd.leadWitnessOrder[rd.currentLeadIndex].spendingPub, *witID)
}

func (rd round_data) getWeight(w *witness, maxBlockValidationPercent float64, maxStaking *big.Int, maxVotes int, maxElected uint64) *big.Float {
	avgStakingAmount := float64(math.GetPercent(rd.valueTotals[string(w.spendingPub)], maxStaking))
	avgBlockValidationPercent := float64(w.validationPercent()) / float64(maxBlockValidationPercent)
	avgVoterCount := float64(len(rd.votes[string(w.spendingPub)])) / float64(maxVotes)
	var avgElectedCount float64
	if maxElected == 0 {
		avgElectedCount = 0
	} else {
		avgElectedCount = float64(w.timesElected) / float64(maxElected)
	}

	return utils.VRFCutoff(avgStakingAmount, avgBlockValidationPercent, avgVoterCount, avgElectedCount)

}

// find the largest values from a list of witnesses.
func (rd round_data) getMaxes() (maxBlockValidationPercent float64, maxStakingAmount *big.Int, maxVoterCount int, maxElectedCount uint64) {
	maxStakingAmount = big.NewInt(0)
	maxBlockValidationPercent = 0.0
	maxVoterCount = 0
	maxElectedCount = 0
	for _, w := range rd.eligibleWitnesses {
		if t := w.validationPercent(); maxBlockValidationPercent < t {
			maxBlockValidationPercent = t
		}

		if t := rd.valueTotals[string(w.spendingPub)]; maxStakingAmount.Cmp(t) == -1 {
			maxStakingAmount = t
		}

		if maxVoterCount < len(rd.votes[string(w.spendingPub)]) {
			maxVoterCount = len(rd.votes[string(w.spendingPub)])
		}

		if maxElectedCount < w.timesElected {
			maxElectedCount = w.timesElected
		}
	}
	return
}
