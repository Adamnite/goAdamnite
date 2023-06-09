package consensus

import (
	"fmt"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"
)

const (
	maxWitnessNumber = 27
)

// TODO: change these to follow the white paper
var maxTimePerRound = time.Minute * 10
var maxBlocksPerRound = 1024

type WitnessInfo struct {
	address common.Address
	voters  []utils.Voter
}

var (
	ErrNotBNode               = fmt.Errorf("node is not setup to handle VM based operations")
	ErrNotANode               = fmt.Errorf("node is not setup to handle transaction based operations")
	ErrCandidateNotApplicable = fmt.Errorf("the candidate reviewed is not applicable for this pools recordings")
	ErrVoteUnVerified         = fmt.Errorf("this vote could not be verified")
)
