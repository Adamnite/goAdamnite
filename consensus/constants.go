package consensus

import (
	"fmt"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"
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
	voters  []utils.Voter
}

var (
	ErrNotBNode = fmt.Errorf("node is not setup to handle VM based operations")
	ErrNotANode = fmt.Errorf("node is not setup to handle transaction based operations")
)
