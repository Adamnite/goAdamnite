package utils

import (
	"math/big"
)

var ( //not actually saved as constants, since it makes most sense to have them as big floats from the start
	stakingAmountWeight          = big.NewFloat(15)
	blockValidationPercentWeight = big.NewFloat(20)
	voterCountWeight             = big.NewFloat(10)
	electedCountWeight           = big.NewFloat(10)
)

// just gets the total weight for a witness' VRF. Is the sum of all weights
func getWeightTotal() *big.Float {
	sum := big.NewFloat(0).Add(stakingAmountWeight, blockValidationPercentWeight)
	sum.Add(sum, voterCountWeight)
	sum.Add(sum, electedCountWeight)
	return sum
}

func VRF(stakingAmount float64, blockValidationPercent float64, voterCount float64, electedCount float64) *big.Float {
	bStakingAmount := big.NewFloat(stakingAmount)
	bBlockValidationPercent := big.NewFloat(blockValidationPercent)
	bVoterCount := big.NewFloat(voterCount)
	bElectedCount := big.NewFloat(electedCount)

	bWeightedAmount := big.NewFloat(0).Mul(bStakingAmount, stakingAmountWeight)
	bWeightedBlockValidationPercent := big.NewFloat(0).Mul(bBlockValidationPercent, blockValidationPercentWeight)
	bWeightedVoterCount := big.NewFloat(0).Mul(bVoterCount, voterCountWeight)
	bWeightedElectedCount := big.NewFloat(0).Mul(bElectedCount, electedCountWeight)

	bDenominator := getWeightTotal()

	weightFloat := big.NewFloat(0)
	weightFloat.Add(weightFloat, bWeightedAmount)
	weightFloat.Add(weightFloat, bWeightedBlockValidationPercent)
	weightFloat.Add(weightFloat, bWeightedVoterCount)
	weightFloat.Add(weightFloat, bWeightedElectedCount)
	return weightFloat.Quo(weightFloat, bDenominator)
}
