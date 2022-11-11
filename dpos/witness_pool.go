package dpos

import (
	cmath "math"
	"math/big"
	"sort"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/params"
)

const (
	StakingAmountWeight          = 15
	BlockValidationPercentWeight = 20
	VoterCountWeight             = 10
	ElectedCountWeight           = 10
)

// const AccuracyMultiple = big.NewFloat(0).

// VRF is an Verifiable Random Function which calculate the weight of the witness.
func VRF(stakingAmount float64, blockValidationPercent float64, voterCount float64, electedCount float64) *big.Float {
	bStakingAmount := big.NewFloat(stakingAmount)
	bBlockValidationPercent := big.NewFloat(blockValidationPercent)
	bVoterCount := big.NewFloat(voterCount)
	bElectedCount := big.NewFloat(electedCount)

	bWeightedAmount := big.NewFloat(0).Mul(bStakingAmount, big.NewFloat(StakingAmountWeight))
	bWeightedBlockValidationPercent := big.NewFloat(0).Mul(bBlockValidationPercent, big.NewFloat(BlockValidationPercentWeight))
	bWeightedVoterCount := big.NewFloat(0).Mul(bVoterCount, big.NewFloat(VoterCountWeight))
	bWeightedElectedCount := big.NewFloat(0).Mul(bElectedCount, big.NewFloat(ElectedCountWeight))

	bDenominator := big.NewFloat(0)
	bDenominator.Add(bDenominator, big.NewFloat(StakingAmountWeight))
	bDenominator.Add(bDenominator, big.NewFloat(BlockValidationPercentWeight))
	bDenominator.Add(bDenominator, big.NewFloat(VoterCountWeight))
	bDenominator.Add(bDenominator, big.NewFloat(ElectedCountWeight))

	weightFloat := big.NewFloat(0)
	weightFloat.Add(weightFloat, bWeightedAmount)
	weightFloat.Add(weightFloat, bWeightedBlockValidationPercent)
	weightFloat.Add(weightFloat, bWeightedVoterCount)
	weightFloat.Add(weightFloat, bWeightedElectedCount)

	return weightFloat
}

type WitnessInfo struct {
	address common.Address
	voters  []types.Voter
}
type WitnessConfig struct {
	WitnessCount uint32 // The total numbers of witness on top tier
}

var DefaultWitnessConfig = WitnessConfig{
	WitnessCount: 27,
}

var DefaultDemoWitnessConfig = WitnessConfig{
	WitnessCount: 3,
}

var WitnessList = []WitnessInfo{{
	address: common.HexToAddress("0x5d8124bb42734acb442b6992c73ecad2651612cd"),
	voters: []types.Voter{
		{
			Address:       common.HexToAddress("0x5117dd7283175dfd686757784de62197bd2179a2"),
			StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(100)),
		},
	},
},
	{
		address: common.HexToAddress("0x5117dd7283175dfd686757784de62197bd2179a2"),
		voters: []types.Voter{
			{
				Address:       common.HexToAddress("0x5d8124bb42734acb442b6992c73ecad2651612cd"),
				StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(50)),
			},
		},
	}}

type WitnessCandidatePool struct {
	config      WitnessConfig
	chainConfig *params.ChainConfig

	witnessCandidates []types.Witness
	// vrfWeights        []float32
	vrfMaps map[string]types.Witness

	selectedWitnesses []types.Witness
	seed              []byte
}

type WitnessPool struct {
	witnesses []types.Witness
	blacklist []types.Witness
}

func (wp *WitnessCandidatePool) GetCandidates() []types.Witness {
	witnessCount := wp.config.WitnessCount
	trustedWitnessCount := witnessCount/3*2 + 1

	var (
		maxStakingAmount          big.Int
		maxBlockValidationPercent float64
		maxVoterCount             int
		maxElectedCount           uint64
		vrfWeights                []float64
		vrfMaps                   map[float64]types.Witness
		witnesses                 []types.Witness
	)

	vrfMaps = make(map[float64]types.Witness)
	maxStakingAmount = *big.NewInt(0)
	maxBlockValidationPercent = 0.0
	maxVoterCount = 0
	maxElectedCount = 0

	for _, w := range wp.witnessCandidates {
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

	for _, w := range wp.witnessCandidates {
		avgStakingAmount := float64(math.GetPercent(w.GetStakingAmount(), &maxStakingAmount))
		avgBlockValidationPercent := float64(w.GetBlockValidationPercents()) / float64(maxBlockValidationPercent)
		avgVoterCount := float64(len(w.GetVoters())) / float64(maxVoterCount)
		avgElectedCount := float64(w.GetElectedCount()) / float64(maxElectedCount)
		weight := VRF(avgStakingAmount, avgBlockValidationPercent, avgVoterCount, avgElectedCount)
		weightVal, _ := weight.Float64()
		vrfWeights = append(vrfWeights, weightVal)
		vrfMaps[weightVal] = w
	}

	sort.Slice(vrfWeights[:], func(i, j int) bool {
		return vrfWeights[i] > vrfWeights[j]
	})

	for i := 0; i < int(cmath.Min(float64(len(vrfWeights)), float64(trustedWitnessCount))); i++ {
		witnesses = append(witnesses, vrfMaps[vrfWeights[i]])
	}

	return witnesses
}

func NewWitnessPool(config WitnessConfig, chainConfig *params.ChainConfig) *WitnessCandidatePool {

	pool := &WitnessCandidatePool{
		config:            config,
		chainConfig:       chainConfig,
		vrfMaps:           make(map[string]types.Witness, 0),
		witnessCandidates: make([]types.Witness, 0),
	}

	if chainConfig.ChainID == params.TestnetChainConfig.ChainID {

		for _, w := range WitnessList {
			witness := &types.WitnessImpl{
				Address: w.address,
				Voters:  w.voters,
			}
			pool.witnessCandidates = append(pool.witnessCandidates, witness)
		}
	}

	var (
		maxStakingAmount          big.Int
		maxBlockValidationPercent float64
		maxVoterCount             int
		maxElectedCount           uint64
	)

	maxStakingAmount = *big.NewInt(0)
	maxBlockValidationPercent = 0.0
	maxVoterCount = 0
	maxElectedCount = 0

	for _, w := range pool.witnessCandidates {
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

	for _, w := range pool.witnessCandidates {

		avgStakingAmount := float64(math.GetPercent(w.GetStakingAmount(), &maxStakingAmount))
		avgBlockValidationPercent := w.GetBlockValidationPercents() / maxBlockValidationPercent
		avgVoterCount := float64(len(w.GetVoters())) / float64(maxVoterCount)
		avgElectedCount := float64(w.GetElectedCount()) / float64(maxElectedCount)
		w.SetWeight(VRF(avgStakingAmount, avgBlockValidationPercent, avgVoterCount, avgElectedCount))
		pool.vrfMaps[string(w.GetPubKey())] = w
	}

	return pool
}

func (cp *WitnessCandidatePool) SetWitnessCandidates(witnessCandidates []types.Witness) {

	cp.witnessCandidates = witnessCandidates
}

func (cp *WitnessCandidatePool) IsTrustedWitness(pubKey crypto.PublicKey, vrfValue []byte, proof []byte) bool {
	//check that the value given is accurate
	if !pubKey.Verify(cp.seed, vrfValue, proof) {
		//if this causes a false return, assume malicious attempt.
		return false
	}

	//converts the bigFloat to an integer between 0-(2^256 - 1). Or a uInt256.
	witnessTesting := cp.vrfMaps[string(pubKey)]
	fWeight := witnessTesting.GetWeight()
	exp := fWeight.MantExp(nil)
	fWeight.SetMantExp(fWeight, 257-exp)
	iWeight, _ := fWeight.Int(nil)

	if iWeight.Cmp(big.NewInt(0)) != 0 { //iWeight == 0
		iWeight.Sub(iWeight, big.NewInt(1)) //needs to subtract one so its always between
	}

	//convert the vrfValue to a big int between 0-(2^256 - 1).  Or a uInt256.
	vrfIValue := new(big.Int)
	vrfIValue.SetBytes(vrfValue)
	if vrfIValue.Cmp(iWeight) == -1 { //vrfIValue<iWeight
		//Check if there is actually space
		betterFit, betterFitPoint := cp._IsBetterFit(witnessTesting)
		if len(cp.selectedWitnesses) < int(cp.config.WitnessCount) {
			cp.selectedWitnesses = append(cp.selectedWitnesses, witnessTesting)
		} else if betterFit {
			//see if they might just fit into the extra bit of space
			cp.selectedWitnesses[betterFitPoint] = witnessTesting
		} else {
			return false
		}

		return true
	}
	return false
}

// returns true, and the index to replace if new witness is a better fit. If false, -1 is the index returned
func (cp *WitnessCandidatePool) _IsBetterFit(newWit types.Witness) (bool, int) {
	//sort everything to be from smallest value to lowest, then compare. So the smallest weight is still most likely
	//to be replaced
	sort.Slice(cp.selectedWitnesses[:], func(i, j int) bool {
		return cp.selectedWitnesses[i].GetWeight().Cmp(cp.selectedWitnesses[j].GetWeight()) == -1
	}) //orders the witnesses by their weight
	//unsure if the witnesses will need to be ordered.
	for i, witness := range cp.selectedWitnesses {
		if witness.GetWeight().Cmp(newWit.GetWeight()) == -1 { //witness.weight < newWit.weight
			return true, i
		}
	}
	return false, -1
}
func (cp *WitnessCandidatePool) GetWitnessPool() *WitnessPool {
	wp := &WitnessPool{
		witnesses: cp.witnessCandidates,
	}
	return wp
}

func (wp *WitnessPool) GetCurrentWitnessAddress(prevWitnessAddr *common.Address) common.Address {
	if prevWitnessAddr == nil {
		return wp.witnesses[0].GetAddress()
	}

	for i, witness := range wp.witnesses {
		if witness.GetAddress() == *prevWitnessAddr {
			if i >= len(wp.witnesses)-1 {
				return wp.witnesses[0].GetAddress()
			} else {
				return wp.witnesses[i+1].GetAddress()
			}
		}
	}
	return common.Address{}
}
