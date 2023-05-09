package consensus

import (
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
)

type consensusHandlingTypes int8

const (
	NetworkingOnly        consensusHandlingTypes = iota
	PrimaryTransactions                          //representing chamber A, or main transactions
	SecondaryTransactions                        //representing chamber B, or VM consensus
)

const (
	prefixKeyOfWitnessPool = "witnesspool-" //whats this for?
	maxWitnessNumber       = 27
)

type WitnessInfo struct {
	address common.Address //TODO: review that this is the correct info for a witness.
	voters  []blockchain.Voter
}
