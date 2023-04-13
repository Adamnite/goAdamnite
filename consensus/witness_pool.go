package dpos

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"sort"
	"time"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/crypto"
)

type Witness struct {
	Address    string
	Stake      uint64
	Reputation uint64
	NumTimes   uint64
}

type VoteResult struct {
	Candidate   string
	TotalVotes  uint64
	VotePercent float64
}

type WitnessPool struct {
	AllWitnesses []*Witness
	VoteResults  map[string]*VoteResult
}

type VRFResult struct {
	Output []byte
	Proof  []byte
}

func (wp *WitnessPool) SelectWitnesses(seed []byte) []*Witness {
	// Sort candidates by number of votes
	candidates := make([]*VoteResult, 0, len(wp.VoteResults))
	for candidate, result := range wp.VoteResults {
		candidates = append(candidates, &VoteResult{Candidate: candidate, TotalVotes: result.TotalVotes})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].TotalVotes > candidates[j].TotalVotes
	})

	// Select top 10% (or at least 28) of candidates for witness pool A
	poolSize := len(candidates) / 10
	if poolSize < 28 {
		poolSize = 28
	}
	poolA := candidates[:poolSize]

	// Calculate weights and scores for each witness in pool A
	poolB := make([]*Witness, 0, 27)
	for _, candidate := range poolA {
		witness := wp.findWitnessByAddress(candidate.Candidate)
		if witness == nil {
			continue
		}

		// Calculate weight using VRF
		vrfResult := wp.calculateVRF(seed, witness)
		if !wp.checkVRF(vrfResult, seed, witness) {
			continue
		}

		// Calculate score based on percentage of votes, reputation, and number of times selected
		score := float64(candidate.TotalVotes) / float64(wp.getTotalVotes()) * float64(witness.Reputation)
		score += float64(witness.NumTimes)
		score *= float64(witness.Stake)

		if wp.checkScore(vrfResult.Output, score) {
			poolB = append(poolB, witness)
		}
		if len(poolB) == 27 {
			break
		}
	}
	return poolB
}

func (wp *WitnessPool) getTotalVotes() uint64 {
	var totalVotes uint64
	for _, result := range wp.VoteResults {
		totalVotes += result.TotalVotes
	}
	return totalVotes
}

func (wp *WitnessPool) findWitnessByAddress(address string) *Witness {
	for _, witness := range wp.AllWitnesses {
		if witness.Address == address {
			return witness
		}
	}
	return nil
}

func (wp *WitnessPool) calculateVRF(seed []byte, witness *Witness) *VRFResult {
	// Generate random nonce
	nonce, _ := rand.Int(rand.Reader, big.NewInt(1<<32-1))

	// Compute VRF
	hash := sha256.Sum256(append(seed, nonce.Bytes()...))
	output := sha256.Sum256(hash[:])
	proof := append(seed, nonce.Bytes()...)

	return &VRFResult{Output: output[:], Proof: proof}
}

func (wp *WitnessPool) checkVRF(result *VRFResult, seed []byte, witness *Witness) bool {
	// Check that proof is valid
	hash := sha256.Sum256(append(seed, result.Proof[4:]...))
	output := sha256.Sum256(hash[:])
	proof := append(seed, result.Proof[4:]...)
	if !wp.verifyProof(proof, output[:], witness.Address) {
	return false
	}
	// Check that output is less than target score
targetScore := wp.calculateTargetScore(seed, witness)
outputInt := new(big.Int).SetBytes(result.Output)
if outputInt.Cmp(targetScore) >= 0 {
	return false
}

return true

}

func (wp *WitnessPool) verifyProof(proof []byte, output []byte, address string) bool {
	// Implement code to verify proof for given address
	// ...
	return true
	}
	
	func (wp *WitnessPool) calculateTargetScore(seed []byte, witness *Witness) *big.Int {
	// Implement code to calculate target score for given witness
	// ...
	return big.NewInt(0)
	}
	
	func (wp *WitnessPool) checkScore(output []byte, score float64) bool {
	// Implement code to check if output is less than or equal to score
	// ...
	return true
	}

	unc (wp *WitnessPool) saveWitnessPool(db adamnitedb.Database) error {
		blob, err := msgpack.Marshal(wp)
		if err != nil {
			return err
		}
		return db.Insert(append([]byte(prefixKeyOfWitnessPool), wp.Hash[:]...), blob)
	}
	
	func (wp *WitnessPool) isVoted(voterAddr common.Address) bool {
		return wp.Votes[voterAddr] != nil
	}
	
	func (wp *WitnessPool) getVoteNum(addr common.Address) *big.Int {
		voteNum := big.NewInt(0)
		if wp.Votes[addr] != nil {
			voteNum.Set(wp.Votes[addr].StakingAmount)
		}
		return voteNum
	}
	
	func (wp *WitnessPool) GetCurrentWitnessAddress(prevWitnessAddr *common.Address) common.Address {
		if prevWitnessAddr == nil {
			if wp.Witnesses == nil || len(wp.Witnesses) == 0 {
				for _, w := range WitnessList {
					witness := &types.WitnessImpl{
						Address: w.address,
						Voters:  w.voters,
					}
					wp.Witnesses = append(wp.Witnesses, witness)
				}
			}
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
		return common.Address{}
	}
	
	
	
