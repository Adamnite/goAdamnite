package utils

import (
	"crypto"
	"math/big"

<<<<<<< Updated upstream
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/math"
=======
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/math"
>>>>>>> Stashed changes
	"github.com/adamnite/go-adamnite/params"
	lru "github.com/hashicorp/golang-lru"
)

type Witness interface {
	// GetAddress returns the witness address
	GetAddress() utils.Address

	// GetVoters returns the list of voters
	GetVoters() []Voter

	SetVoters(voters []Voter)

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
	Address        utils.Address
	Voters         []Voter
	Prove          []byte
	WeightVRF      *big.Float
	PubKey         crypto.PublicKey
	blocksReviewed uint64
	blocksApproved uint64
}

func (w *WitnessImpl) GetAddress() utils.Address {
	return w.Address
}

func (w *WitnessImpl) GetVoters() []Voter {
	return w.Voters
}
func (w *WitnessImpl) SetVoters(voters []Voter) {
	w.Voters = voters
}
func (w *WitnessImpl) GetBlockValidationPercents() float64 {
	if w.blocksReviewed == 0 {
		return 0.5
	}
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

type WitnessPool struct {
	chainConfig *params.ChainConfig

	witnessCandidates []Witness
	// vrfWeights        []float32
<<<<<<< Updated upstream
	vrfMaps map[common.Address]*Witness //map the account address to the witness pointer (no need to re-save it)

	Witnesses []Witness
	seed      []byte
	Votes     map[common.Address]*Voter
	sigcache  *lru.ARCCache
	Number    uint64
	Hash      common.Hash
=======
	vrfMaps map[utils.Address]*Witness //map the account address to the witness pointer (no need to re-save it)

	Witnesses []Witness
	seed      []byte
	Votes     map[utils.Address]*Voter
	sigcache  *lru.ARCCache
	Number    uint64
	Hash      utils.Hash
>>>>>>> Stashed changes
	blacklist []Witness
}

func NewWitnessPool(chainConfig *params.ChainConfig) *WitnessPool {

	pool := &WitnessPool{
		chainConfig: chainConfig,

<<<<<<< Updated upstream
		vrfMaps:           make(map[common.Address]*Witness),
		witnessCandidates: []Witness{},

		Votes: map[common.Address]*Voter{},
=======
		vrfMaps:           make(map[utils.Address]*Witness),
		witnessCandidates: []Witness{},

		Votes: map[utils.Address]*Voter{},
>>>>>>> Stashed changes
	}

	if chainConfig.ChainID == params.TestnetChainConfig.ChainID {

		// for _, w := range DefaultTestnetGenesisBlock().WitnessList {
		// 	witness := &WitnessImpl{
		// 		Address: w.address,
		// 		Voters:  w.voters,
		// 	}
		// 	pool.Witnesses = append(pool.Witnesses, witness)
		// }

		vrfMaps, _ := SetVRFItems(pool.Witnesses)
		for _, w := range vrfMaps {
			pool.vrfMaps[w.GetAddress()] = &w
		}

	}
	return pool
}

func SetVRFItems(witnesses []Witness) (vrfMaps map[float64]Witness, vrfWeights []float64) {
	vrfMaps, vrfWeights = make(map[float64]Witness), []float64{} //variable assignment for clarity.

	maxBlockValidationPercent, maxStakingAmount, maxVoterCount, maxElectedCount := getMaxesFrom(witnesses)
	for _, w := range witnesses {
		avgStakingAmount := float64(math.GetPercent(w.GetStakingAmount(), &maxStakingAmount))
		avgBlockValidationPercent := float64(w.GetBlockValidationPercents()) / float64(maxBlockValidationPercent)
		avgVoterCount := float64(len(w.GetVoters())) / float64(maxVoterCount)
		avgElectedCount := float64(w.GetElectedCount()) / float64(maxElectedCount)
		weight := VRF(avgStakingAmount, avgBlockValidationPercent, avgVoterCount, avgElectedCount)
		w.SetWeight(weight)
		weightVal, _ := weight.Float64()
		vrfWeights = append(vrfWeights, weightVal)
		vrfMaps[weightVal] = w
	}
	return
}

// find the largest values from a list of witnesses.
func getMaxesFrom(witnesses []Witness) (maxBlockValidationPercent float64, maxStakingAmount big.Int, maxVoterCount int, maxElectedCount uint64) {
	maxStakingAmount = *big.NewInt(0)
	maxBlockValidationPercent = 0.0
	maxVoterCount = 0
	maxElectedCount = 0
	for _, w := range witnesses {
		if t := w.GetBlockValidationPercents(); maxBlockValidationPercent < t {
			maxBlockValidationPercent = t
		}

		if t := w.GetStakingAmount(); maxStakingAmount.Cmp(t) == -1 {
			maxStakingAmount = *t
		}

		if maxVoterCount < len(w.GetVoters()) {
			maxVoterCount = len(w.GetVoters())
		}

		if maxElectedCount < w.GetElectedCount() {
			maxElectedCount = w.GetElectedCount()
		}
	}
	return
}
