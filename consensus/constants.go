package consensus

import (
	"fmt"
	"time"
)

const (
	maxWitnessNumber = 27
)

// TODO: change these to follow the white paper
var maxTimePerRound = time.Minute * 10
var maxTimePrecision = time.Second * 2
var maxBlocksPerRound uint64 = 27 * 6
var maxTransactionsPerBlock int = 255

var (
	ErrNotBNode               = fmt.Errorf("node is not setup to handle VM based operations")
	ErrNotANode               = fmt.Errorf("node is not setup to handle transaction based operations")
	ErrCandidateNotApplicable = fmt.Errorf("the candidate reviewed is not applicable for this pools recordings")
	ErrVoteUnVerified         = fmt.Errorf("this vote could not be verified")
)
