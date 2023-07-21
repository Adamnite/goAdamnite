package poh

import (
	"errors"
	"io"
	cmath "math"
	"math/big"
	"sort"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/math"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/params"
	"github.com/adamnite/go-adamnite/utils"
	lru "github.com/hashicorp/golang-lru"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	StakingAmountWeight          = 15 //%27.27...
	BlockValidationPercentWeight = 20 //%49.09...
	VoterCountWeight             = 10 //%18.18...
	ElectedCountWeight           = 10 //%18.18...
	prefixKeyOfDBWitnessPool     = "db-witnesspool-"
	maxWitnessNumber             = 18

	blockInterval       = 1
	EpochBlockCount     = 162
	inMemorySignatures  = 4096
	checkpointInterval  = 1024
	inMemoryWitnessPool = 128
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
	address utils.Address
	voters  []utils.Voter
}

type DBWitnessConfig struct {
	WitnessCount uint32 // The total numbers of witness on top tier
}

var DefaultDBWitnessConfig = DBWitnessConfig{
	WitnessCount: 18,
}

var WitnessList = []DBWitnessInfo{{
	address: utils.HexToAddress("3HCiFhyA1Kv3s25BeABHt7wW6N8y"),
	voters: []utils.Voter{
		{
			// Address:       utils.HexToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"),
			StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(100)),
		},
	},
},
	{
		address: utils.HexToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"),
		voters: []utils.Voter{
			{
				// Address:       utils.HexToAddress("3HCiFhyA1Kv3s25BeABHt7wW6N8y"),
				StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(50)),
			},
		},
	}}

var (
	// When a list of signers for a block is requested, errunknownblock is returned.
	// This is not part of the local blockchain.
	errUnknownBlock = errors.New("unknown block")

	//If the timestamp of the block is lower than errInvalidTimestamp
	//Timestamp of the previous block + minimum block period.
	ErrInvalidTimestamp         = errors.New("invalid timestamp")
	ErrWaitForPrevBlock         = errors.New("wait for last block arrived")
	ErrApplyNextBlock           = errors.New("apply the next block")
	ErrMismatchSignerAndWitness = errors.New("mismatch block signer and witness")
	ErrInvalidWitness           = errors.New("invalid witness")
	ErrInvalidApplyBlockTime    = errors.New("invalid time to apply the block")
	ErrNilBlockHeader           = errors.New("nil block header returned")
	ErrUnknownAncestor          = errors.New("unknown ancestor")
	ErrInvalidVotingChain       = errors.New("invalid voting chain")
)

type PoHData struct {
	Witnesses []utils.Witness `json:"dbwitnesses"`

	Votes map[utils.Address]utils.Voter `json:"dbvotes"`
}

type DBWitnessPool struct {
	config      DBWitnessConfig
	chainConfig *params.ChainConfig

	witnessCandidates []utils.Witness
	// vrfWeights        []float32
	vrfMaps map[string]utils.Witness

	Witnesses []utils.Witness
	seed      []byte
	Votes     map[utils.Address]*utils.Voter
	sigcache  *lru.ARCCache
	Number    uint64
	Hash      utils.Hash
}

func NewDBRoundWitnessPool(config DBWitnessConfig, chainConfig *params.ChainConfig, sigcache *lru.ARCCache, number uint64, hash utils.Hash, witnesses []utils.Witness) *DBWitnessPool {

	pool := &DBWitnessPool{
		config:            config,
		chainConfig:       chainConfig,
		sigcache:          sigcache,
		Number:            number,
		Hash:              hash,
		vrfMaps:           make(map[string]utils.Witness, 0),
		witnessCandidates: make([]utils.Witness, 0),
		Witnesses:         witnesses,
		Votes:             map[utils.Address]*utils.Voter{},
	}
	if chainConfig.ChainID == params.TestnetChainConfig.ChainID {
		if number == 0 {
			for _, w := range WitnessList {
				witness := &utils.WitnessImpl{
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
				pool.vrfMaps[w.GetAddress().String()] = w
			}
		}

	}
	return pool
}

func NewDBWitnessPool(config DBWitnessConfig, chainConfig *params.ChainConfig) *DBWitnessPool {

	pool := &DBWitnessPool{
		config:      config,
		chainConfig: chainConfig,

		vrfMaps:           make(map[string]utils.Witness, 0),
		witnessCandidates: make([]utils.Witness, 0),

		Votes: map[utils.Address]*utils.Voter{},
	}
	if chainConfig.ChainID == params.TestnetChainConfig.ChainID {

		for _, w := range WitnessList {
			witness := &utils.WitnessImpl{
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
			pool.vrfMaps[w.GetAddress().String()] = w
		}

	}
	return pool
}

func (wp *DBWitnessPool) CalcWitnesses() []utils.Witness {
	witnessCount := wp.config.WitnessCount
	trustedWitnessCount := witnessCount/3*2 + 1

	var (
		maxStakingAmount          big.Int
		maxBlockValidationPercent float64
		maxVoterCount             int
		maxElectedCount           uint64
		vrfWeights                []float64
		vrfMaps                   map[float64]utils.Witness
		witnesses                 []utils.Witness
	)

	vrfMaps = make(map[float64]utils.Witness)
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

func (cp *DBWitnessPool) SetWitnessCandidates(witnessCandidates []utils.Witness) {

	cp.witnessCandidates = witnessCandidates
}

func GetDBWitnessPool(config *params.ChainConfig, sigcache *lru.ARCCache, db adamnitedb.Database, hash utils.Hash) (*DBWitnessPool, error) {

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

func (wp *DBWitnessPool) SaveDBWitnessPool(db adamnitedb.Database) error {
	blob, err := msgpack.Marshal(wp)
	if err != nil {
		return err
	}
	return db.Insert(append([]byte(prefixKeyOfDBWitnessPool), wp.Hash[:]...), blob)
}

func (wp *DBWitnessPool) IsVoted(voterAddr utils.Address) bool {
	return wp.Votes[voterAddr] != nil
}
func (wp *DBWitnessPool) copy() *DBWitnessPool {
	cpy := &DBWitnessPool{
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

func (wp *DBWitnessPool) DbWitnessPoolFromBlockHeader(headers []*types.BlockHeader) (*DBWitnessPool, error) {
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

		witness, err := Ecrecover(header, witnesspool.sigcache)
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

		pohData := PoHData{}
		PohDataDecode(header.Extra, &pohData)
		if number%EpochBlockCount == 0 {
			if number > 0 {

				witnesspool.Votes = make(map[utils.Address]*utils.Voter)
				witnesspool.witnessCandidates = make([]utils.Witness, 0)
			}
			witnesspool.Witnesses = pohData.Witnesses
		}
		votes := pohData.Votes
		for sender, vote := range votes {

			witnesspool.Votes[sender] = &utils.Voter{
				// Address:       vote.Address,
				StakingAmount: vote.StakingAmount,
			}
			count := 0
			// for _, wpCandidate := range witnesspool.witnessCandidates {
			// 	// if wpCandidate.GetAddress() == vote.Address {
			// 	// 	tmpVoters := append(wpCandidate.GetVoters(), utils.Voter{Address: vote.Address, StakingAmount: vote.StakingAmount})
			// 	// 	wpCandidate.SetVoters(tmpVoters)
			// 	// 	count++
			// 	// }
			// }

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

func PohDataDecode(bytes []byte, data *PoHData) error {
	return msgpack.Unmarshal(bytes, &data)
}
func PohDataEncode(poh PoHData) []byte {
	bytes, _ := msgpack.Marshal(poh)
	return bytes
}
func Ecrecover(header *types.BlockHeader, sigcache *lru.ARCCache) (utils.Address, error) {

	// If the signature's already cached, return that
	hash := header.Hash()
	if address, known := sigcache.Get(hash); known {
		return address.(utils.Address), nil
	}

	signature := header.Extra

	// Recover the public key and the Adamnite address
	pubkey, err := crypto.Recover(SealHash(header).Bytes(), signature)
	if err != nil {
		return utils.Address{}, err
	}

	signer := crypto.PubkeyByteToAddress(pubkey)

	sigcache.Add(hash, signer)
	return signer, nil
}

func SealHash(header *types.BlockHeader) (hash utils.Hash) {
	hasher := crypto.NewRipemd160State()
	encodeSignHeader(hasher, header)
	hasher.Sum(hash[:0])
	return hash
}

func encodeSignHeader(w io.Writer, header *types.BlockHeader) {
	err := msgpack.NewEncoder(w).Encode([]interface{}{
		header.ParentHash,
		header.Witness,
		header.WitnessRoot,
		// header.CurrentEpoch,
		header.Number,
		header.Signature,
		header.StateRoot,
		header.Extra,
		header.Time,
		header.TransactionRoot,
		header.DBWitness,
	})
	if err != nil {
		panic("can't encode: " + err.Error())
	}
}
