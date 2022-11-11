package dpos

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/log15"
)

type EpochContext struct {
	TimeStamp int64
	DposEnv   types.DposEnv
	statedb   *statedb.StateDB
}

func (ec *EpochContext) countVotes() (votes map[common.Address]*big.Int, err error) {
	votes = map[common.Address]*big.Int{}

	witnessTrie := ec.DposEnv.WitnessTrie()
	witnessCandidateTrie := ec.DposEnv.WitnessCandidateTrie()
	statedb := ec.statedb

	iterCandidate := trie.NewIterator(witnessCandidateTrie.NodeIterator(nil))
	existCandidate := iterCandidate.Next()
	if !existCandidate {
		return votes, errors.New("no witness candidates")
	}

	for existCandidate {
		candidate := iterCandidate.Value
		candidateAddr := common.BytesToAddress(candidate)
		witnessIterator := trie.NewIterator(witnessTrie.PrefixIterator(candidate))
		existWitness := witnessIterator.Next()
		if !existWitness {
			votes[candidateAddr] = new(big.Int)
			existCandidate = iterCandidate.Next()
			continue
		}
		for existWitness {
			witness := witnessIterator.Value
			score, ok := votes[candidateAddr]
			if !ok {
				score = new(big.Int)
			}
			witnessAddr := common.BytesToAddress(witness)

			weight := statedb.GetOrNewStateObj(witnessAddr).Balance()
			score.Add(score, weight)
			votes[candidateAddr] = score
			existWitness = witnessIterator.Next()
		}
		existCandidate = iterCandidate.Next()
	}
	return votes, nil
}

func (ec *EpochContext) tryElect(genesis, parent *types.BlockHeader, witnessCandidatePool WitnessCandidatePool) error {

	genesisEpoch := int64(genesis.Time / EpochBlockCount)
	prevEpoch := int64(parent.Time / EpochBlockCount)
	currentEpoch := ec.TimeStamp / EpochBlockCount

	prevEpochIsGenesis := prevEpoch == genesisEpoch
	if prevEpochIsGenesis && prevEpoch < currentEpoch {
		prevEpoch = currentEpoch - 1
	}

	prevEpochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(prevEpochBytes, uint64(prevEpoch))
	iter := trie.NewIterator(ec.DposEnv.MintCntTrie().PrefixIterator(prevEpochBytes))

	for i := prevEpoch; i < currentEpoch; i++ {

		if !prevEpochIsGenesis && iter.Next() {
			if err := ec.removeWitness(prevEpoch, genesis); err != nil {
				return err
			}
		}

		votes, err := ec.countVotes()
		if err != nil {
			return err
		}

		voters := make([]types.Voter, 0)
		for addr, weight := range votes {
			voters = append(voters, types.Voter{addr, weight})
		}

		candidates := sortableAddresses{}

		addresses := make([]common.Address, 0, len(votes))
		for addr := range votes {
			addresses = append(addresses, addr)
		}
		sort.Slice(addresses, func(i, j int) bool {
			return votes[addresses[i]].Cmp(votes[addresses[j]]) > 0
		})

		lenIndex := 0
		for _, addr := range addresses {
			if len(votes) > 10 {
				if lenIndex <= len(votes)/10 {
					candidates = append(candidates, &sortableAddress{addr, votes[addr]})
					lenIndex++
				}
			} else {
				candidates = append(candidates, &sortableAddress{addr, votes[addr]})
			}
		}
		maxWitnessSize := 27
		if len(candidates) < 27 {
			maxWitnessSize = len(candidates)
		}
		safeSize := maxWitnessSize*2/3 + 1

		if len(candidates) < safeSize {
			return errors.New("too few candidates")
		}
		sort.Sort(candidates)
		if len(candidates) > maxWitnessSize {
			candidates = candidates[:maxWitnessSize]
		}

		witnessCandidates := make([]types.Witness, 0)

		for _, candidate := range candidates {
			witness := &types.WitnessImpl{
				Address: candidate.address,
				Voters:  voters,
			}
			witnessCandidates = append(witnessCandidates, witness)
		}

		witnessCandidatePool.SetWitnessCandidates(witnessCandidates)

		epochTrie, _ := types.NewEpochTrie(common.Hash{}, ec.DposEnv.DB())
		ec.DposEnv.SetEpoch(epochTrie)
		ec.DposEnv.SetWitnesses(witnessCandidatePool.GetCandidates())
		log15.Info("Come to new epoch", "prevEpoch", i, "nextEpoch", i+1)
	}
	return nil
}

func (ec *EpochContext) removeWitness(epoch int64, genesis *types.BlockHeader) error {
	witnesses, err := ec.DposEnv.GetWitnesses()

	maxWitnessSize := 27
	safeSize := int(maxWitnessSize*2/3 + 1)

	if err != nil {
		return fmt.Errorf("failed to get witness: %s", err)
	}
	if len(witnesses) == 0 {
		return errors.New("no witness could be removed")
	}

	epochDuration := int64(164 * 10)
	blockInterval := genesis.Time

	timeOfFirstBlock := int64(0)

	if ec.TimeStamp-0 < 164 {
		epochDuration = (ec.TimeStamp - timeOfFirstBlock)
	}

	needRemoveWitnesses := sortableAddresses{}
	for _, witness := range witnesses {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(epoch))
		key = append(key, []byte(witness.GetAddress().String())...)
		cnt := int64(0)
		if cntBytes := ec.DposEnv.MintCntTrie().Get(key); cntBytes != nil {
			cnt = int64(binary.BigEndian.Uint64(cntBytes))
		}

		if cnt < epochDuration/int64(blockInterval)/int64(maxWitnessSize)/2 {

			needRemoveWitnesses = append(needRemoveWitnesses, &sortableAddress{witness.GetAddress(), big.NewInt(cnt)})
		}
	}

	needRemoveWitnessCnt := len(needRemoveWitnesses)
	if needRemoveWitnessCnt <= 0 {
		return nil
	}
	sort.Sort(sort.Reverse(needRemoveWitnesses))

	candidateCount := 0
	iter := trie.NewIterator(ec.DposEnv.WitnessCandidateTrie().NodeIterator(nil))
	for iter.Next() {
		candidateCount++
		if candidateCount >= needRemoveWitnessCnt+int(safeSize) {
			break
		}
	}

	for i, witness := range needRemoveWitnesses {

		if candidateCount <= int(safeSize) {
			log15.Info("No more witness candidate can be removed", "prevEpochID", epoch, "witnessCandidateCount", candidateCount, "needRemoveCount", len(needRemoveWitnesses)-i)
			return nil
		}

		if err := ec.DposEnv.RemoveWitnessCandidate(witness.address); err != nil {
			return err
		}
		//
		candidateCount--
		log15.Info("remove witness candidate", "prevEpochID", epoch, "witness candidate", witness.address.String(), "mintCnt", witness.weight.String())
	}
	return nil
}

func (ec *EpochContext) lookupWitness(now int64, blockInterval uint64) (witness common.Address, err error) {
	witness = common.Address{}
	offset := now % EpochBlockCount
	if offset%int64(blockInterval) != 0 {
		return common.Address{}, errors.New("invalid time to make the block")
	}
	offset /= int64(blockInterval)

	witnesses, err := ec.DposEnv.GetWitnesses()
	if err != nil {
		return common.Address{}, err
	}
	witnessSize := len(witnesses)
	if witnessSize == 0 {
		return common.Address{}, errors.New("failed to lookup witness")
	}
	offset %= int64(witnessSize)
	return witnesses[offset].GetAddress(), nil
}

type sortableAddress struct {
	address common.Address
	weight  *big.Int
}
type sortableAddresses []*sortableAddress

func (p sortableAddresses) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p sortableAddresses) Len() int      { return len(p) }
func (p sortableAddresses) Less(i, j int) bool {
	if p[i].weight.Cmp(p[j].weight) < 0 {
		return false
	} else if p[i].weight.Cmp(p[j].weight) > 0 {
		return true
	} else {
		return p[i].address.String() < p[j].address.String()
	}
}
