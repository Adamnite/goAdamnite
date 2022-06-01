package core

import (
	cmath "math"
	"math/big"
	"sort"

	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/core/types"
)

const (
	StakingAmountWeight          = 15
	BlockValidationPercentWeight = 20
	VoterCountWeight             = 10
	ElectedCountWeight           = 10
)

// VRF is an Verfiable Random Function which calcuate the weight of the witness.
func VRF(stakingAmount float32, blockValidationPercent float32, voterCount float32, electedCount float32) float32 {
	weight := (stakingAmount*StakingAmountWeight + blockValidationPercent*BlockValidationPercentWeight + voterCount*VoterCountWeight + electedCount*ElectedCountWeight) / (StakingAmountWeight + BlockValidationPercentWeight + VoterCountWeight + ElectedCountWeight)
	return weight
}

func ChooseWitnesses(witnessPool WitnessCandidatePool) []types.Witness {
	witnessCount := witnessPool.config.WitnessCount
	trustedWitnessCount := witnessCount/3*2 + 1

	var (
		maxStakingAmount          big.Int
		maxBlockValidationPercent float32
		maxVoterCount             int
		maxElectedCount           uint64
		vrfWeights                []float32
		vrfMaps                   map[float32]types.Witness
		witnesses                 []types.Witness
	)

	vrfMaps = make(map[float32]types.Witness)
	maxStakingAmount = *big.NewInt(0)
	maxBlockValidationPercent = 0.0
	maxVoterCount = 0
	maxElectedCount = 0

	for _, w := range witnessPool.witnessCandidates {
		if maxBlockValidationPercent < w.GetBlockValidationPercents() {
			maxBlockValidationPercent = w.GetBlockValidationPercents()
		}

		if maxStakingAmount.Cmp(w.GetStakingAmount()) == -1 {
			maxStakingAmount = *w.GetStakingAmount()
		}

		if maxVoterCount < len(w.GetVoters()) {
			maxVoterCount = len(w.GetVoters())
		}

		if maxElectedCount < w.GetElectedCount() {
			maxElectedCount = w.GetElectedCount()
		}
	}

	for _, w := range witnessPool.witnessCandidates {
		avgStakingAmount := math.GetPercent(w.GetStakingAmount(), &maxStakingAmount)
		avgBlockValidationPercent := float32(w.GetBlockValidationPercents()) / float32(maxBlockValidationPercent)
		avgVoterCount := float32(len(w.GetVoters())) / float32(maxVoterCount)
		avgElectedCount := float32(w.GetElectedCount()) / float32(maxElectedCount)
		weight := VRF(avgStakingAmount, avgBlockValidationPercent, avgVoterCount, avgElectedCount)
		vrfWeights = append(vrfWeights, weight)
		vrfMaps[weight] = w
	}

	sort.Slice(vrfWeights[:], func(i, j int) bool {
		return vrfWeights[i] > vrfWeights[j]
	})

	for i := 0; i < int(cmath.Min(float64(len(vrfWeights)), float64(trustedWitnessCount))); i++ {
		witnesses = append(witnesses, vrfMaps[vrfWeights[i]])
	}

	// ToDo: We need to select witnesses randomly

	return witnesses
}
