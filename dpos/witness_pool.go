package dpos

import (
	cmath "math"
	"math/big"
	"sort"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/dpos/poh"
	"github.com/adamnite/go-adamnite/params"
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

type WitnessConfig struct {
	WitnessCount uint32 // The total numbers of witness on top tier
}

var DefaultWitnessConfig = WitnessConfig{
	WitnessCount: 27,
}

var DefaultDemoWitnessConfig = WitnessConfig{
	WitnessCount: 3,
}

type WitnessPool struct {
	config      WitnessConfig
	chainConfig *params.ChainConfig

	witnessCandidates types.WitnessList
	// vrfWeights        []float32
	vrfMaps map[string]types.Witness

	witnesses types.WitnessList
	seed      []byte
	Votes     map[common.Address]*types.Voter
	sigcache  *lru.ARCCache
	Number    uint64
	Hash      common.Hash
	blacklist types.WitnessList

	// Blockchain Information
	epochNum  uint64
	blockNum  uint64
}

func NewRoundWitnessPool(config WitnessConfig, chainConfig *params.ChainConfig, sigcache *lru.ARCCache, number uint64, hash common.Hash, witnesses []types.Witness) *WitnessPool {

	pool := &WitnessPool{
		config:            config,
		chainConfig:       chainConfig,
		sigcache:          sigcache,
		Number:            number,
		Hash:              hash,
		vrfMaps:           make(map[string]types.Witness, 0),
		witnessCandidates: make([]types.Witness, 0),
		witnesses:         witnesses,
		Votes:             map[common.Address]*types.Voter{},
	}

	return pool
}

func NewWitnessPool(config *WitnessConfig, chainConfig *params.ChainConfig, witnessList []types.Witness) *WitnessPool {

	pool := &WitnessPool{
		config:      *config,
		chainConfig: chainConfig,

		vrfMaps:           make(map[string]types.Witness, 0),
		witnessCandidates: make([]types.Witness, 0),

		Votes: map[common.Address]*types.Voter{},
	}

	for _, w := range witnessList {
		witness := &types.WitnessImpl{
			Address: w.GetAddress(),
			Voters:  w.GetVoters(),
		}
		pool.witnesses = append(pool.witnesses, witness)
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

	for _, w := range pool.witnesses {
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

	for _, w := range pool.witnesses {

		avgStakingAmount := float64(math.GetPercent(w.GetStakingAmount(), &maxStakingAmount))
		avgBlockValidationPercent := w.GetBlockValidationPercents() / maxBlockValidationPercent
		avgVoterCount := float64(len(w.GetVoters())) / float64(maxVoterCount)
		avgElectedCount := float64(w.GetElectedCount()) / float64(maxElectedCount)
		w.SetWeight(VRF(avgStakingAmount, avgBlockValidationPercent, avgVoterCount, avgElectedCount))
		pool.vrfMaps[string(w.GetPubKey())] = w
	}

	return pool
}

func (wp *WitnessPool) CalcWitnesses() []types.Witness {
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

func (cp *WitnessPool) SetWitnessCandidates(witnessCandidates []types.Witness) {

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
		if len(cp.witnesses) < int(cp.config.WitnessCount) {
			cp.witnesses = append(cp.witnesses, witnessTesting)
		} else if betterFit {
			//see if they might just fit into the extra bit of space
			cp.witnesses[betterFitPoint] = witnessTesting
		} else {
			return false
		}

		return true
	}
	return false
}

// returns true, and the index to replace if new witness is a better fit. If false, -1 is the index returned
func (cp *WitnessPool) _IsBetterFit(newWit types.Witness) (bool, int) {
	//sort everything to be from smallest value to lowest, then compare. So the smallest weight is still most likely
	//to be replaced
	sort.Slice(cp.witnesses[:], func(i, j int) bool {
		return cp.witnesses[i].GetWeight().Cmp(cp.witnesses[j].GetWeight()) == -1
	}) //orders the witnesses by their weight
	//unsure if the witnesses will need to be ordered.
	for i, witness := range cp.witnesses {
		if witness.GetWeight().Cmp(newWit.GetWeight()) == -1 { //witness.weight < newWit.weight
			return true, i
		}
	}
	return false, -1
}
func GetWitnessPool(config *params.ChainConfig, sigcache *lru.ARCCache, db adamnitedb.Database, hash common.Hash) (*WitnessPool, error) {

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

func (wp *WitnessPool) SaveWitnessPool(db adamnitedb.Database) error{
	witnessList := wp.WitnessList();
	rawdb.WriteWitnessList(db, wp.EpochNum(), witnessList)

	return nil
}

func LoadWitnessPool(config *WitnessConfig, chainConfig *params.ChainConfig, db adamnitedb.AdamniteDBReader) (*WitnessPool, error) {
	epochNum := rawdb.ReadEpochNumber(db)
	wList, err := rawdb.ReadWitnessList(db, epochNum)
	if err != nil {
		return nil, err
	}

	wp := NewWitnessPool(config, chainConfig, wList)
	return wp, nil
}

func (wp *WitnessPool) IsVoted(voterAddr common.Address) bool {
	return wp.Votes[voterAddr] != nil
}

func (wp *WitnessPool) getVoteNum(addr common.Address) *big.Int {
	voteNum := big.NewInt(0)
	if wp.Votes[addr] != nil {
		voteNum.Set(wp.Votes[addr].StakingAmount)
	}
	return voteNum
}

func (wp *WitnessPool) copy() *WitnessPool {
	cpy := &WitnessPool{
		config:            wp.config,
		sigcache:          wp.sigcache,
		Number:            wp.Number,
		Hash:              wp.Hash,
		witnesses:         wp.witnesses,
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
		for index, wp_witness := range wp.witnesses {
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

				witnesspool.Votes = make(map[common.Address]*types.Voter)
				witnesspool.witnessCandidates = make([]types.Witness, 0)
			}
			witnesspool.witnesses = dposData.Witnesses
		}
		votes := dposData.Votes
		for sender, vote := range votes {

			witnesspool.Votes[sender] = &types.Voter{
				Address:       vote.Address,
				StakingAmount: vote.StakingAmount,
			}
			count := 0
			for _, wpCandidate := range witnesspool.witnessCandidates {
				if wpCandidate.GetAddress() == vote.Address {
					tmpVoters := append(wpCandidate.GetVoters(), types.Voter{Address: vote.Address, StakingAmount: vote.StakingAmount})
					wpCandidate.SetVoters(tmpVoters)
					count++
				}
			}

			if count == 0 {
				tmpVotes := make([]types.Voter, 0)
				tmpVotes = append(tmpVotes, types.Voter{Address: vote.Address, StakingAmount: vote.StakingAmount})
				tmpWitness := &types.WitnessImpl{
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

func (wp *WitnessPool) WitnessList() []types.Witness {
	return wp.witnesses
}

func (wp *WitnessPool) BlackList() []types.Witness {
	return wp.blacklist
}

func (wp *WitnessPool) RootHash() common.Hash {
	return types.DeriveSha(wp.witnesses, trie.NewStackTrie(nil))
}

func (wp *WitnessPool) BlockNum() uint64 { return wp.blockNum }

func (wp *WitnessPool) EpochNum() uint64 { return wp.epochNum }
