package types

import (
	"math/big"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
)

type Witness interface {
	// GetAddress returns the witness address
	GetAddress() common.Address

	// GetVoters returns the list of voters
	GetVoters() []Voter

	// GetBlockValidationPercents returns the percent of
	GetBlockValidationPercents() float64

	// GetElectedCount returns the number of elected round
	GetElectedCount() uint64

	// GetStakingAmount returns the total amount of staking for vote
	GetStakingAmount() *big.Int

	SetWeight(weight *big.Float)

	GetWeight() *big.Float

	GetPubKey() crypto.PublicKey

	BlockReviewed(bool)
}

type WitnessImpl struct {
	Address        common.Address
	Voters         []Voter
	Prove          []byte
	WeightVRF      *big.Float
	PubKey         crypto.PublicKey
	blocksReviewed uint64
	blocksApproved uint64
}

func (w *WitnessImpl) GetAddress() common.Address {
	return w.Address
}

func (w *WitnessImpl) GetVoters() []Voter {
	return w.Voters
}

func (w *WitnessImpl) GetBlockValidationPercents() float64 {
	return float64(w.blocksApproved) / float64(w.blocksReviewed)
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

func (w *WitnessImpl) GetWeight() *big.Float {
	return w.WeightVRF
}

func (w *WitnessImpl) SetWeight(weight *big.Float) {
	w.WeightVRF = weight
}
func (w *WitnessImpl) GetPubKey() crypto.PublicKey {
	return w.PubKey
}

func (w *WitnessImpl) BlockReviewed(successful bool) {
	w.blocksReviewed++
	if successful {
		w.blocksApproved++
	}
}
