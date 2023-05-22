package rpc

import (
	"github.com/adamnite/go-adamnite/utils"
	encoding "github.com/vmihailenco/msgpack/v5"
)

// server items on the RPC side for consensus only

func (a *AdamniteServer) SetConsensusHandlers(
	newCandidate func(utils.Candidate) error,
	newVote func(utils.Voter) error) {
	a.newCandidateHandler = newCandidate
	a.newVoteHandler = newVote
}

const NewCandidateEndpoint = "AdamniteServer.NewCandidate"

func (a *AdamniteServer) NewCandidate(params *[]byte, reply *[]byte) error {
	a.print("New Candidate")
	if a.newCandidateHandler == nil {
		a.print("not set to handle candidates")
		return nil //we aren't setup to handle it, just forward it
	}
	var candidateProposed utils.Candidate
	if err := encoding.Unmarshal(*params, &candidateProposed); err != nil {
		a.printError("New Candidate", err)
		return err
	}
	return a.newCandidateHandler(candidateProposed)
}

const NewVoteEndpoint = "AdamniteServer.NewVote"

func (a *AdamniteServer) NewVote(params *[]byte, reply *[]byte) error {
	a.print("New Vote")
	if a.newVoteHandler == nil {
		a.print("not set to handle candidates")
		return nil //we aren't setup to handle it, just forward it
	}
	var vote utils.Voter
	if err := encoding.Unmarshal(*params, &vote); err != nil {
		a.printError("New Vote", err)
		return err
	}
	return a.newVoteHandler(vote)
}
