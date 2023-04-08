
import (
	"errors"
	"fmt"
	"sync"
	"https://github.com/adamnite/go-adamnite/crypto"
	"https://github.com/adamnite/go-adamnite/common"
    "https://github.com/adamnite/go-adamnite/core/types"

)

type VoteTracker struct {
	mu        sync.Mutex
	candidates map[Address]Candidate
}

func NewVoteTracker(candidates map[Address]Candidate) *VoteTracker {
	return &VoteTracker{
		candidates: candidates,
	}
}

func (vt *VoteTracker) HandleVote(vote Vote) error {
	vt.mu.Lock()
	defer vt.mu.Unlock()

	candidate, ok := vt.candidates[vote.Header.Recipient]
	if !ok {
		return errors.New("recipient is not a candidate")
	}

	if candidate.Active {
		if vote.Header.Timestamp > candidate.Deadline {
			return errors.New("vote is past deadline")
		}

		if err := candidate.VerifyVote(vote); err != nil {
			return err
		}

		candidate.Votes += vote.Header.Amount
	}

	return nil
}

func (vt *VoteTracker) HandleEvent(event Event) {
	vt.mu.Lock()
	defer vt.mu.Unlock()

	switch event.Type {
	case EventRoundStarted:
		for _, candidate := range vt.candidates {
			candidate.ResetVotes()
		}
	case EventCandidateActivated:
		candidate, ok := vt.candidates[event.Address]
		if !ok {
			return
		}

		candidate.Active = true
		candidate.Deadline = event.Timestamp + candidate.RoundDuration()
	}
}

func (vt *voteTracker) CheckValidity(vote *Vote) error {
    // Verify the vote using the sender's public key
    sender := vt.gossip.getPubKey(vote.Header.Sender)
    err := vote.Verify(sender)
    if err != nil {
        return fmt.Errorf("invalid vote: %v", err)
    }
    
    // Check that the recipient is a valid candidate
    recipient := vote.Header.Recipient
    _, ok := vt.candidates[recipient]
    if !ok {
        return fmt.Errorf("invalid candidate address: %v", recipient)
    }
    
    // Check that the vote is not expired
    if vote.Header.Timestamp+vt.consensus.VoteTimeout < vt.currentHeight {
        return fmt.Errorf("vote has expired")
    }
    
    // Add the vote to the candidate's total
    vt.candidates[recipient].addVote(vote.Header.Amount)
    vt.voteCount[recipient] += 1
    
    return nil
}

type VoteAggregator struct {
	mu        sync.Mutex
	candidates map[Address]Candidate
}

func NewVoteAggregator(candidates map[Address]Candidate) *VoteAggregator {
	return &VoteAggregator{
		candidates: candidates,
	}
}

func (va *VoteAggregator) Aggregate() error {
	va.mu.Lock()
	defer va.mu.Unlock()

	for _, candidate := range va.candidates {
		if !candidate.Active {
			continue
		}

		if candidate.Votes < candidate.RequiredVotes() {
			return fmt.Errorf("candidate %v did not receive enough votes", candidate.Address)
		}

		if err := candidate.UpdateReputation(); err != nil {
			return err
		}
	}

	return nil
}