package utils

import (
	"math/big"

	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

type Candidate struct {
	// round specific data
	Round     uint64 //round number proposing for
	Seed      []byte //should be consistent between all candidates. I think using round N-1's lead signature works best
	VRFValue  []byte
	VRFProof  []byte
	StartTime uint64 //when that round should start
	// us specific data

	VRFKey crypto.PublicKey //our VRF public Key

	ConsensusPool int8 //support type that this is being pitched for
	NetworkString string
	NodeID        crypto.PublicKey

	InitialVote Voter //this is the list of people who voted for this candidate
}

func NewCandidate(
	round uint64, seed []byte, vrfPrivate crypto.PrivateKey,
	startAt uint64, targetPool uint8, netPoint string,
	nodeId crypto.PublicKey, spender accounts.Account, stakeAmount *big.Int,
) (*Candidate, error) {
	pub, _ := vrfPrivate.Public()
	can := Candidate{
		Round:         round,
		Seed:          seed,
		VRFKey:        pub,
		StartTime:     startAt,
		ConsensusPool: int8(targetPool),
		NetworkString: netPoint,
		NodeID:        nodeId,
	}
	can.VRFValue, can.VRFProof = vrfPrivate.Prove(seed)
	v := NewVote(spender.PublicKey, stakeAmount)
	if err := v.SignTo(can, spender); err != nil {
		return nil, err
	}

	can.InitialVote = v
	return &can, nil
}
func (c Candidate) UpdatedCandidate(round uint64, newSeed []byte, vrfPrivate crypto.PrivateKey, startAt uint64, spender accounts.Account) (*Candidate, error) {
	can := Candidate{
		Round:         round,
		Seed:          newSeed,
		VRFKey:        c.VRFKey,
		StartTime:     startAt,
		ConsensusPool: c.ConsensusPool,
		NetworkString: c.NetworkString,
		NodeID:        c.NodeID,
	}
	can.VRFValue, can.VRFProof = vrfPrivate.Prove(newSeed)
	v := NewVote(spender.PublicKey, c.InitialVote.StakingAmount)
	if err := v.SignTo(can, spender); err != nil {
		return nil, err
	}

	can.InitialVote = v
	return &can, nil
}

// get a hash of the candidate, but this does not include votes
func (c *Candidate) Hash() []byte {
	byteForm := []byte{byte(c.Round)}
	byteForm = append(byteForm, c.Seed...)
	byteForm = append(byteForm, byte(c.StartTime)) //TODO: change this to properly format it
	byteForm = append(byteForm, c.VRFKey...)
	byteForm = append(byteForm, c.NodeID...)
	byteForm = append(byteForm, []byte(c.NetworkString)...)
	byteForm = append(byteForm, byte(c.ConsensusPool))
	return crypto.Sha512(byteForm)
}

func (c Candidate) VerifyVote(vote Voter) bool {
	if vote.StakingAmount.Sign() != 1 {
		return false
	}
	spendingAccount := vote.Account()
	candidateHash := c.Hash()

	voteAndCandidateHash := append(candidateHash, vote.Hash()...)
	return spendingAccount.Verify(voteAndCandidateHash, vote.Signature)
}
