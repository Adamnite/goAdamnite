package consensus

import (
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
	witnessesMap      map[string]*witness
	votes             map[string][]*utils.Voter //map[witnessNodeID]->votes for that witness, in that round
	valueTotals       map[string]*big.Int       //map[witnessNodeID] -> total amount staked on them
	vrfValues         map[string][]byte         //map[witnessNodeID]->vrf value
	vrfProofs         map[string][]byte         //map[witnessNodeID]->vrf proof
	vrfCutoffs        map[string]*big.Float     //map[witnessNodeID]->vrf cutoff value
	seed              []byte
	openToApply       bool
	roundStartTime    time.Time
	blocksInRound     uint64
}

func newRoundData(seed []byte) *round_data {
	newRound := round_data{
		witnessesMap:   make(map[string]*witness),
		votes:          make(map[string][]*utils.Voter),
		valueTotals:    make(map[string]*big.Int),
		vrfValues:      make(map[string][]byte),
		vrfProofs:      make(map[string][]byte),
		vrfCutoffs:     make(map[string]*big.Float),
		openToApply:    true,
		seed:           seed,
		roundStartTime: time.Now().UTC(),
	}
	return &newRound
}
func (rd *round_data) getNextRoundSeed() []byte {
	if len(rd.witnesses) == 0 {
		return []byte{}
	}
	return rd.vrfValues[string(rd.witnesses[0].nodeID)]
}
func (rd *round_data) addEligibleWitness(w *witness, vrfVal []byte, vrfProof []byte) {
	//since every candidate needs a vote (from themselves) to run, we can check the vote count
	if rd.votes[string(w.nodeID)] != nil {
		return
	}

	rd.eligibleWitnesses = append(rd.eligibleWitnesses, w)
	rd.vrfValues[string(w.nodeID)] = vrfVal
	rd.vrfProofs[string(w.nodeID)] = vrfProof
	rd.valueTotals[string(w.nodeID)] = big.NewInt(0)
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

// select witnesses should be called only once all votes have been received, and the next round needs to start. Returns witnesses selected
func (rd *round_data) selectWitnesses(goalCount int) ([]*witness, []byte) {
	if rd.openToApply { //close the applications and select the winning witnesses
		rd.openToApply = false
		maxBlockValidationPercent, maxStaking, maxVotes, maxElected := rd.getMaxes()
		for _, w := range rd.eligibleWitnesses {
			weight := rd.getWeight(w, maxBlockValidationPercent, maxStaking, maxVotes, maxElected)
			rd.vrfCutoffs[string(w.nodeID)] = weight
		}
	}

	passingWitnesses := []*witness{}
	witnessVrfValueFloat := make(map[string]*big.Float)
	//32 bytes, all set to 255
	maxVRFVal := big.NewFloat(0).SetInt(big.NewInt(1).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}))
	for _, w := range rd.eligibleWitnesses {
		witnessVRFValue := big.NewFloat(0).Quo(
			big.NewFloat(1).SetInt(big.NewInt(1).SetBytes(rd.vrfValues[string(w.nodeID)])),
			maxVRFVal,
		)
		witnessVrfValueFloat[string(w.nodeID)] = witnessVRFValue
		if rd.vrfCutoffs[string(w.nodeID)].Cmp(witnessVRFValue) != -1 { //the witnesses VRF value is less than or equal to their cutoff
			passingWitnesses = append(passingWitnesses, w)
		}
	}
	if len(rd.eligibleWitnesses) <= goalCount { //we don't have enough witnesses to try and generate more
		passingWitnesses = rd.eligibleWitnesses
	}
	//sort the passing witnesses based on the difference between the cutoff and their score
	sort.Slice(passingWitnesses, func(i, j int) bool {
		//we sort no matter what so that we can get the next rounds seed
		aId := passingWitnesses[i].nodeID
		bId := passingWitnesses[j].nodeID
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
			aId := rd.eligibleWitnesses[i].nodeID
			bId := rd.eligibleWitnesses[j].nodeID
			aDif := big.NewFloat(0).Sub(rd.vrfCutoffs[string(aId)], witnessVrfValueFloat[string(aId)])
			bDif := big.NewFloat(0).Sub(rd.vrfCutoffs[string(bId)], witnessVrfValueFloat[string(bId)])
			return aDif.Cmp(bDif) == 1
		})
		passingWitnesses = append(passingWitnesses,
			rd.eligibleWitnesses[len(passingWitnesses):goalCount]...)
	}

	rd.witnesses = passingWitnesses
	for _, w := range passingWitnesses {
		rd.witnessesMap[string(w.nodeID)] = w
	}
	return passingWitnesses, rd.vrfValues[string(passingWitnesses[0].nodeID)]
}
func (rd round_data) getWeight(w *witness, maxBlockValidationPercent float64, maxStaking *big.Int, maxVotes int, maxElected uint64) *big.Float {
	avgStakingAmount := float64(math.GetPercent(rd.valueTotals[string(w.nodeID)], maxStaking))
	avgBlockValidationPercent := float64(w.validationPercent()) / float64(maxBlockValidationPercent)
	avgVoterCount := float64(len(rd.votes[string(w.nodeID)])) / float64(maxVotes)
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

		if t := rd.valueTotals[string(w.nodeID)]; maxStakingAmount.Cmp(t) == -1 {
			maxStakingAmount = t
		}

		if maxVoterCount < len(rd.votes[string(w.nodeID)]) {
			maxVoterCount = len(rd.votes[string(w.nodeID)])
		}

		if maxElectedCount < w.timesElected {
			maxElectedCount = w.timesElected
		}
	}
	return
}
