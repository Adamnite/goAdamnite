package utils

import (
	"github.com/adamnite/go-adamnite/crypto"
)

type Candidate struct {
	// round specific data
	Round     uint64 //round number proposing for
	Seed      []byte //should be consistent between all candidates. I think using round N-1's lead signature works best
	StartTime uint64 //when that round should start
	// us specific data

	VRFKey crypto.PublicKey //our VRF public Key

	ConsensusPool int8 //support type that this is being pitched for
	NetworkString string
	NodeID        crypto.PublicKey

	InitialVote Voter //this is the list of people who voted for this candidate
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

	voteAndCandidateHash := append(candidateHash, vote.StakingAmount.Bytes()...)
	voteAndCandidateHash = append(voteAndCandidateHash, vote.From...)
	return spendingAccount.Verify(voteAndCandidateHash, vote.Signature)
}
