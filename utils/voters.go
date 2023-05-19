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
	candidateHash := candidate.Hash()
	voteAndCandidateHash := append(candidateHash, v.StakingAmount.Bytes()...)
	voteAndCandidateHash = append(voteAndCandidateHash, v.From...)
	signature, err := signer.Sign(crypto.Sha512(voteAndCandidateHash))
	if err != nil {
		return err
	}
	v.Signature = signature
	return nil
}

func (v Voter) Address() common.Address {
	return accounts.AccountFromPubBytes(v.From).Address
}
func (v Voter) Account() accounts.Account {
	return accounts.AccountFromPubBytes(v.From)
}
