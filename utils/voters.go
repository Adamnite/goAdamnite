package utils

import (
	"math/big"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

type Voter struct {
	To            []byte
	From          []byte
	PoolCategory  uint8
	StakingAmount *big.Int
	Signature     []byte
}

func NewVote(from []byte, amount *big.Int) Voter {
	return Voter{
		From:          from,
		StakingAmount: amount,
	}
}
func (v *Voter) SignTo(candidate Candidate, signer accounts.Account) error {
	v.To = candidate.NodeID
	v.PoolCategory = candidate.ConsensusPool
	candidateHash := candidate.Hash()
	voteAndCandidateHash := append(candidateHash, v.Hash()...)
	signature, err := signer.Sign(voteAndCandidateHash)
	if err != nil {
		return err
	}
	v.Signature = signature
	return nil
}
func (v Voter) Hash() []byte {
	h := append(v.StakingAmount.Bytes(), v.From...)
	h = append(h, v.PoolCategory)
	return crypto.Sha512(h)
}
func (v Voter) Address() common.Address {
	return accounts.AccountFromPubBytes(v.From).Address
}
func (v Voter) Account() accounts.Account {
	return accounts.AccountFromPubBytes(v.From)
}
