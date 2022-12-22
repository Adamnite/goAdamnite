package poh

import (
	cmath "math"
	"math/big"
	"sort"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/params"
	lru "github.com/hashicorp/golang-lru"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	StakingAmountWeight          = 15
	BlockValidationPercentWeight = 20
	VoterCountWeight             = 10
	ElectedCountWeight           = 10
	prefixKeyOfDBWitnessPool     = "db-witnesspool-"
	maxWitnessNumber             = 18
)

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

type DBWitnessInfo struct {
	address common.Address
	voters  []types.Voter
}

type DBWitnessConfig struct {
	WitnessCount uint32 // The total numbers of witness on top tier
}

var DefaultDBWitnessConfig = DBWitnessConfig{
	WitnessCount: 18,
}

var WitnessList = []DBWitnessInfo{{
	address: common.HexToAddress("3HCiFhyA1Kv3s25BeABHt7wW6N8y"),
	voters: []types.Voter{
		{
			Address:       common.HexToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"),
			StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(100)),
		},
	},
},
	{
		address: common.HexToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"),
		voters: []types.Voter{
			{
				Address:       common.HexToAddress("3HCiFhyA1Kv3s25BeABHt7wW6N8y"),
				StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(50)),
			},
		},
	}}

type DBWitnessPool struct {
	config      DBWitnessConfig
	chainConfig *params.ChainConfig

	witnessCandidates []types.Witness
	// vrfWeights        []float32
	vrfMaps map[string]types.Witness

	Witnesses []types.Witness
	seed      []byte
	Votes     map[common.Address]*types.Voter
	sigcache  *lru.ARCCache
	Number    uint64
	Hash      common.Hash
}

func NewDBRoundWitnessPool(config DBWitnessConfig, chainConfig *params.ChainConfig, sigcache *lru.ARCCache, number uint64, hash common.Hash, witnesses []types.Witness) *DBWitnessPool {

	pool := &DBWitnessPool{
		config:            config,
		chainConfig:       chainConfig,
		sigcache:          sigcache,
		Number:            number,
		Hash:              hash,
		vrfMaps:           make(map[string]types.Witness, 0),
		witnessCandidates: make([]types.Witness, 0),
		Witnesses:         witnesses,
		Votes:             map[common.Address]*types.Voter{},
	}
	if chainConfig.ChainID == params.TestnetChainConfig.ChainID {
		if number == 0 {
			for _, w := range WitnessList {
				witness := &types.WitnessImpl{
					Address: w.address,
					Voters:  w.voters,
				}
				pool.Witnesses = append(pool.Witnesses, witness)
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

			for _, w := range pool.Witnesses {
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

			for _, w := range pool.Witnesses {

				avgStakingAmount := float64(math.GetPercent(w.GetStakingAmount(), &maxStakingAmount))
				avgBlockValidationPercent := w.GetBlockValidationPercents() / maxBlockValidationPercent
				avgVoterCount := float64(len(w.GetVoters())) / float64(maxVoterCount)
				avgElectedCount := float64(w.GetElectedCount()) / float64(maxElectedCount)
				w.SetWeight(VRF(avgStakingAmount, avgBlockValidationPercent, avgVoterCount, avgElectedCount))
				pool.vrfMaps[string(w.GetPubKey())] = w
			}
		}

	}
	return pool
}

func NewDBWitnessPool(config DBWitnessConfig, chainConfig *params.ChainConfig) *DBWitnessPool {

	pool := &DBWitnessPool{
		config:      config,
		chainConfig: chainConfig,

		vrfMaps:           make(map[string]types.Witness, 0),
		witnessCandidates: make([]types.Witness, 0),

		Votes: map[common.Address]*types.Voter{},
	}
	if chainConfig.ChainID == params.TestnetChainConfig.ChainID {

		for _, w := range WitnessList {
			witness := &types.WitnessImpl{
				Address: w.address,
				Voters:  w.voters,
			}
			pool.Witnesses = append(pool.Witnesses, witness)
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

		for _, w := range pool.Witnesses {
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

		for _, w := range pool.Witnesses {

			avgStakingAmount := float64(math.GetPercent(w.GetStakingAmount(), &maxStakingAmount))
			avgBlockValidationPercent := w.GetBlockValidationPercents() / maxBlockValidationPercent
			avgVoterCount := float64(len(w.GetVoters())) / float64(maxVoterCount)
			avgElectedCount := float64(w.GetElectedCount()) / float64(maxElectedCount)
			w.SetWeight(VRF(avgStakingAmount, avgBlockValidationPercent, avgVoterCount, avgElectedCount))
			pool.vrfMaps[string(w.GetPubKey())] = w
		}

	}
	return pool
}

func (wp *DBWitnessPool) CalcWitnesses() []types.Witness {
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

func (cp *DBWitnessPool) SetWitnessCandidates(witnessCandidates []types.Witness) {

	cp.witnessCandidates = witnessCandidates
}

func GetDBWitnessPool(config *params.ChainConfig, sigcache *lru.ARCCache, db adamnitedb.Database, hash common.Hash) (*DBWitnessPool, error) {

	blob, err := db.Get(append([]byte(prefixKeyOfDBWitnessPool), hash[:]...))
	if err != nil {
		return nil, err
	}
	witnessPool := new(DBWitnessPool)
	if err := msgpack.Unmarshal(blob, witnessPool); err != nil {
		return nil, err
	}
	witnessPool.chainConfig = config
	witnessPool.sigcache = sigcache
	return witnessPool, nil

}

func (wp *DBWitnessPool) saveWitnessPool(db adamnitedb.Database) error {
	blob, err := msgpack.Marshal(wp)
	if err != nil {
		return err
	}
	return db.Insert(append([]byte(prefixKeyOfDBWitnessPool), wp.Hash[:]...), blob)
}
