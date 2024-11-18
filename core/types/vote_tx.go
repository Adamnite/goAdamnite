package types

import (
	"math/big"

	"github.com/adamnite/go-adamnite/common"
)

type VoteTransaction struct {
	Type      TxType         // transaction type
	Nonce     uint64         // nonce of the sender account
	Candidate common.Address // the candidate address of witness or block producer
	AtePrice  *big.Int       // wei per gas
	AteMax    uint64         // gas limit
	V, R, S   *big.Int       // signature value

}

func NewVoteTransaction(nonce uint64, candidate common.Address, atePrice *big.Int, ateMax uint64) *Transaction {
	return NewTx(&VoteTransaction{
		Type:      VOTE_TX,
		Nonce:     nonce,
		Candidate: candidate,
		AtePrice:  atePrice,
		AteMax:    ateMax,
	})
}

func (tx *VoteTransaction) copy() Transaction_Data {
	cpy := &VoteTransaction{
		Type:      tx.Type,
		Nonce:     tx.Nonce,
		Candidate: tx.Candidate,
		AtePrice:  new(big.Int),
		AteMax:    tx.AteMax,
		V:         new(big.Int),
		R:         new(big.Int),
		S:         new(big.Int),
	}

	if tx.AtePrice != nil {
		cpy.AtePrice.Set(tx.AtePrice)
	}

	if tx.V != nil {
		cpy.V.Set(tx.V)
	}

	if tx.R != nil {
		cpy.R.Set(tx.R)
	}

	if tx.S != nil {
		cpy.S.Set(tx.S)
	}

	return cpy
}

func (tx *VoteTransaction) txtype() TxType         { return VOTE_TX }
func (tx *VoteTransaction) chain_TYPE() *big.Int   { return nil }
func (tx *VoteTransaction) amount() *big.Int       { return nil }
func (tx *VoteTransaction) message() []byte        { return nil }
func (tx *VoteTransaction) message_size() *big.Int { return nil }
func (tx *VoteTransaction) ATE_MAX() uint64        { return tx.AteMax }
func (tx *VoteTransaction) ATE_price() *big.Int    { return tx.AtePrice }
func (tx *VoteTransaction) to() *common.Address    { return nil }
func (tx *VoteTransaction) nonce() uint64          { return tx.Nonce }

func (tx *VoteTransaction) rawSignature() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *VoteTransaction) setSignatue(chainID, v, r, s *big.Int) {
	tx.V, tx.R, tx.S = v, r, s
}
