package consensus

import (
	"fmt"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"
)

const (
	maxWitnessNumber = 27
)

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
