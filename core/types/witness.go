package types

import (
	"math/big"

	"github.com/adamnite/go-adamnite/common"
)

type Witness interface {
	// GetAddress returns the witness address
	GetAddress() common.Address

	// GetVoters returns the list of voters
	GetVoters() []Voter

	// GetBlockValidationPercents returns the percent of
	GetBlockValidationPercents() float32

	// GetElectedCount returns the number of elected round
	GetElectedCount() uint64

	// GetStakingAmount returns the total amount of staking for vote
	GetStakingAmount() *big.Int
}

type WitnessImpl struct {
	Address   common.Address
	Voters    []Voter
	Prove     []byte
	WeightVRF []byte
}

func (w *WitnessImpl) GetAddress() common.Address {
	return w.Address
}

func (w *WitnessImpl) GetVoters() []Voter {
	return w.Voters
}

func (w *WitnessImpl) GetBlockValidationPercents() float32 {
	return 0.9
}

func (w *WitnessImpl) GetElectedCount() uint64 {
	return 1
}

func (w *WitnessImpl) GetStakingAmount() *big.Int {
	totalStakingAmount := big.NewInt(0)

	for _, w := range w.Voters {
		totalStakingAmount = new(big.Int).Add(totalStakingAmount, w.StakingAmount)
	}

	return totalStakingAmount
}
