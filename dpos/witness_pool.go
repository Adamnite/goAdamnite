package dpos

import (
	cmath "math"
	"math/big"
	"sort"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/dpos/poh"
	"github.com/adamnite/go-adamnite/params"
	"github.com/adamnite/go-adamnite/utils"
	lru "github.com/hashicorp/golang-lru"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	StakingAmountWeight          = 15
	BlockValidationPercentWeight = 20
	VoterCountWeight             = 10
	ElectedCountWeight           = 10
	prefixKeyOfWitnessPool       = "witnesspool-"
	maxWitnessNumber             = 27
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
	address bytes.Address
	voters  []utils.Voter
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

// var WitnessList = []WitnessInfo{{
// 	address: common.HexToAddress("3HCiFhyA1Kv3s25BeABHt7wW6N8y"),
// 	voters: []utils.Voter{
// 		{
// 			Address:       common.HexToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"),
// 			StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(100)),
// 		},
// 	},
// },
// 	{
// 		address: common.HexToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"),
// 		voters: []utils.Voter{
// 			{
// 				Address:       common.HexToAddress("3HCiFhyA1Kv3s25BeABHt7wW6N8y"),
// 				StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(50)),
// 			},
// 		},
// 	}}

type WitnessPool struct {
	config      WitnessConfig
	chainConfig *params.ChainConfig

	witnessCandidates []utils.Witness
	// vrfWeights        []float32
	vrfMaps map[string]utils.Witness

	Witnesses []utils.Witness
	seed      []byte
	Votes     map[bytes.Address]*utils.Voter
	sigcache  *lru.ARCCache
	Number    uint64
	Hash      bytes.Hash
	blacklist []utils.Witness
}

func NewRoundWitnessPool(config WitnessConfig, chainConfig *params.ChainConfig, sigcache *lru.ARCCache, number uint64, hash bytes.Hash, witnesses []utils.Witness) *WitnessPool {

	pool := &WitnessPool{
		config:            config,
		chainConfig:       chainConfig,
		sigcache:          sigcache,
		Number:            number,
		Hash:              hash,
		vrfMaps:           make(map[string]utils.Witness, 0),
		witnessCandidates: make([]utils.Witness, 0),
		Witnesses:         witnesses,
		Votes:             map[bytes.Address]*utils.Voter{},
	}

	return pool
}

func NewWitnessPool(config WitnessConfig, chainConfig *params.ChainConfig) *WitnessPool {

	pool := &WitnessPool{
		config:      config,
		chainConfig: chainConfig,

		vrfMaps:           make(map[string]utils.Witness, 0),
		witnessCandidates: make([]utils.Witness, 0),

		Votes: map[bytes.Address]*utils.Voter{},
	}

	if chainConfig.ChainID == params.TestnetChainConfig.ChainID {

		vrfMaps, _ := setVRFItems(pool.Witnesses)
		for _, w := range vrfMaps {
			pool.vrfMaps[w.GetAddress().String()] = w
		}

	}

	return pool
}
func getMaxesFrom(witnesses []utils.Witness) (maxBlockValidationPercent float64, maxStakingAmount big.Int, maxVoterCount int, maxElectedCount uint64) {
	maxStakingAmount = *big.NewInt(0)
	maxBlockValidationPercent = 0.0
	maxVoterCount = 0
	maxElectedCount = 0
	for _, w := range witnesses {
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
	return
}

func (wp *WitnessPool) CalcWitnesses() []utils.Witness {
	witnessCount := wp.config.WitnessCount
	trustedWitnessCount := witnessCount/3*2 + 1

	var (
		witnesses []utils.Witness
	)
	vrfMaps, vrfWeights := setVRFItems(wp.witnessCandidates)

	sort.Slice(vrfWeights[:], func(i, j int) bool {
		return vrfWeights[i] > vrfWeights[j]
	})

	for i := 0; i < int(cmath.Min(float64(len(vrfWeights)), float64(trustedWitnessCount))); i++ {
		witnesses = append(witnesses, vrfMaps[vrfWeights[i]])
	}

	return witnesses
}
func setVRFItems(witnesses []utils.Witness) (vrfMaps map[float64]utils.Witness, vrfWeights []float64) {
	vrfMaps, vrfWeights = make(map[float64]utils.Witness), []float64{} //variable assignment for clarity.

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

func (cp *WitnessPool) SetWitnessCandidates(witnessCandidates []utils.Witness) {

	cp.witnessCandidates = witnessCandidates
}

func (cp *WitnessPool) IsTrustedWitness(pubKey crypto.PublicKey, vrfValue []byte, proof []byte) bool {
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
		if len(cp.Witnesses) < int(cp.config.WitnessCount) {
			cp.Witnesses = append(cp.Witnesses, witnessTesting)
		} else if betterFit {
			//see if they might just fit into the extra bit of space
			cp.Witnesses[betterFitPoint] = witnessTesting
		} else {
			return false
		}

		return true
	}
	return false
}

// returns true, and the index to replace if new witness is a better fit. If false, -1 is the index returned
func (cp *WitnessPool) _IsBetterFit(newWit utils.Witness) (bool, int) {
	//sort everything to be from smallest value to lowest, then compare. So the smallest weight is still most likely
	//to be replaced
	sort.Slice(cp.Witnesses[:], func(i, j int) bool {
		return cp.Witnesses[i].GetWeight().Cmp(cp.Witnesses[j].GetWeight()) == -1
	}) //orders the witnesses by their weight
	//unsure if the witnesses will need to be ordered.
	for i, witness := range cp.Witnesses {
		if witness.GetWeight().Cmp(newWit.GetWeight()) == -1 { //witness.weight < newWit.weight
			return true, i
		}
	}
	return false, -1
}
func GetWitnessPool(config *params.ChainConfig, sigcache *lru.ARCCache, db adamnitedb.Database, hash bytes.Hash) (*WitnessPool, error) {

	blob, err := db.Get(append([]byte(prefixKeyOfWitnessPool), hash[:]...))
	if err != nil {
		return nil, err
	}
	witnessPool := new(WitnessPool)
	if err := msgpack.Unmarshal(blob, witnessPool); err != nil {
		return nil, err
	}
	witnessPool.chainConfig = config
	witnessPool.sigcache = sigcache
	return witnessPool, nil

}

func (wp *WitnessPool) saveWitnessPool(db adamnitedb.Database) error {
	blob, err := msgpack.Marshal(wp)
	if err != nil {
		return err
	}
	return db.Insert(append([]byte(prefixKeyOfWitnessPool), wp.Hash[:]...), blob)
}

func (wp *WitnessPool) isVoted(voterAddr bytes.Address) bool {
	return wp.Votes[voterAddr] != nil
}

func (wp *WitnessPool) getVoteNum(addr bytes.Address) *big.Int {
	voteNum := big.NewInt(0)
	if wp.Votes[addr] != nil {
		voteNum.Set(wp.Votes[addr].StakingAmount)
	}
	return voteNum
}

func (wp *WitnessPool) GetCurrentWitnessAddress(prevWitnessAddr *bytes.Address) bytes.Address {
	if prevWitnessAddr == nil {
		return wp.Witnesses[0].GetAddress()

	}

	for i, witness := range wp.Witnesses {
		if witness.GetAddress() == *prevWitnessAddr {
			if i >= len(wp.Witnesses)-1 {
				return wp.Witnesses[0].GetAddress()
			} else {
				return wp.Witnesses[i+1].GetAddress()
			}
		}
	}
	return bytes.Address{}
}

func (wp *WitnessPool) copy() *WitnessPool {
	cpy := &WitnessPool{
		config:            wp.config,
		sigcache:          wp.sigcache,
		Number:            wp.Number,
		Hash:              wp.Hash,
		Witnesses:         wp.Witnesses,
		Votes:             wp.Votes,
		witnessCandidates: wp.witnessCandidates,
	}
	return cpy
}

func (wp *WitnessPool) witnessPoolFromBlockHeader(headers []*types.BlockHeader) (*WitnessPool, error) {
	if len(headers) == 0 {
		return wp, nil
	}
	for i := 0; i < len(headers)-1; i++ {
		if headers[i].Number.Uint64()+1 != headers[i+1].Number.Uint64() {
			return nil, ErrInvalidVotingChain
		}
	}

	if headers[0].Number.Uint64() != wp.Number+1 {
		return nil, ErrInvalidVotingChain
	}

	witnesspool := wp.copy()
	for _, header := range headers {
		number := header.Number.Uint64()

		witness, err := poh.Ecrecover(header, witnesspool.sigcache)
		if err != nil {
			return nil, err
		}
		i := 0
		for index, wp_witness := range wp.Witnesses {
			println(index)
			if wp_witness.GetAddress() == witness {
				i++
			}
		}

		if i == 0 {
			return nil, ErrMismatchSignerAndWitness

		}

		dposData := DposData{}
		DposDataDecode(header.Extra, &dposData)
		if number%EpochBlockCount == 0 {
			if number > 0 {

				witnesspool.Votes = make(map[bytes.Address]*utils.Voter)
				witnesspool.witnessCandidates = make([]utils.Witness, 0)
			}
			witnesspool.Witnesses = dposData.Witnesses
		}
		votes := dposData.Votes
		for sender, vote := range votes {

			witnesspool.Votes[sender] = &utils.Voter{
				// Address:       vote.Address,
				StakingAmount: vote.StakingAmount,
			}
			count := 0

			if count == 0 {
				tmpVotes := make([]utils.Voter, 0)
				// tmpVotes = append(tmpVotes, utils.Voter{Address: vote.Address, StakingAmount: vote.StakingAmount})
				tmpWitness := &utils.WitnessImpl{
					Address: sender,
					Voters:  tmpVotes,
				}
				witnesspool.witnessCandidates = append(witnesspool.witnessCandidates, tmpWitness)
			}

		}

	}

	witnesspool.Number += uint64(len(headers))
	witnesspool.Hash = headers[len(headers)-1].Hash()
	return witnesspool, nil
}
