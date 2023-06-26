package consensus

import (
	"bytes"
	"fmt"
	"log"
	"math/big"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

// The node that can handle consensus systems.
// Is built on top of a networking node, using a netNode to handle the networking interactions
type ConsensusNode struct {
	thisCandidate        *utils.Candidate
	currentRound         uint64
	votesSeen            map[string][]*utils.Voter   //candidate nodeID->voters array
	candidates           map[string]*utils.Candidate //candidate nodeID->Candidate(just sorting by nodeID)
	candidateStakeValues map[string]*big.Int         //using a mapping to keep track of how much has been staked into each candidate
	netLogic             *networking.NetNode
	handlingType         networking.NetworkTopLayerType

	spendingKey   accounts.Account // each consensus node is forced to have its own account to spend from.
	participation accounts.Account
	vrfKey        accounts.Account
	state         *statedb.StateDB
	chain         *blockchain.Blockchain //we need to keep the chain
	ocdbLink      string                 //off chain database, if running the VM verification, this should be local.
	vm            *VM.Machine

	autoVoteForNode *common.Address
	autoVoteWith    *common.Address
	autoStakeAmount *big.Int
}

func NewAConsensus(spendingKey accounts.Account) (*ConsensusNode, error) {
	//TODO: setup the chain data and whatnot
	conNode, err := newConsensus(nil, nil)
	conNode.spendingKey = spendingKey
	conNode.handlingType = networking.PrimaryTransactions
	return conNode, err
}

func newConsensus(state *statedb.StateDB, chain *blockchain.Blockchain) (*ConsensusNode, error) {
	participation, err := accounts.GenerateAccount()
	if err != nil {
		return nil, err
	}
	hostingNode := networking.NewNetNode(participation.Address)
	con := ConsensusNode{
		netLogic:             hostingNode,
		handlingType:         networking.NetworkingOnly,
		state:                state,
		chain:                chain,
		participation:        *participation,
		votesSeen:            make(map[string][]*utils.Voter),
		candidateStakeValues: make(map[string]*big.Int),
		candidates:           make(map[string]*utils.Candidate),
	}
	if err := hostingNode.AddFullServer(state, chain, con.ReviewTransaction, con.ReviewCandidacy, con.ReviewVote); err != nil {
		log.Printf("error:%v", err)
		return nil, err
	}
	return &con, nil
}
func (con *ConsensusNode) ReviewTransaction(transaction *utils.Transaction) error {
	//TODO: give this a quick look over, review it, if its good, add it locally and propagate it out, otherwise, ignore it.
	return nil
}

// review authenticity of a vote, as well as recording it for our own records. If no errors are returned, will propagate further
func (con *ConsensusNode) ReviewVote(vote utils.Voter) error {
	candidate, exists := con.candidates[string(crypto.PublicKey(vote.To))]
	if !exists {
		return fmt.Errorf("we don't have that account saved, we could throw an error, or check for that candidate first, maybe save unknown votes to check again at the end?")
	}
	//TODO: check the balance of these voters!
	verified := candidate.VerifyVote(vote)
	if !verified {
		return ErrVoteUnVerified
	}
	//assuming by here it is legit.
	con.votesSeen[string(candidate.NodeID)] = append(con.votesSeen[string(candidate.NodeID)], &vote)
	con.candidateStakeValues[string(candidate.NodeID)].Add(con.candidateStakeValues[string(candidate.NodeID)], vote.StakingAmount)
	return nil
}

// review a candidate proposal. The Consensus node may add a vote on. If no errors are returned, assume it is fine to forward along.
func (con *ConsensusNode) ReviewCandidacy(proposed utils.Candidate) error {
	log.Println("reviewing candidate!")
	//review the base matches as we believe it should
	if proposed.ConsensusPool != int8(con.handlingType) ||
		proposed.Round != con.currentRound+1 {
		return ErrCandidateNotApplicable
	}
	//review that the initial vote is signed correctly
	if !proposed.VerifyVote(proposed.InitialVote) {
		log.Println("someone lied in a vote")
		return ErrVoteUnVerified
	}
	con.candidates[string(proposed.NodeID)] = &proposed
	con.votesSeen[string(proposed.NodeID)] = []*utils.Voter{&proposed.InitialVote}
	con.candidateStakeValues[string(proposed.NodeID)] = proposed.InitialVote.StakingAmount

	//if we made it here, this is most likely a viable candidate
	//TODO: check if we have this candidate in our contacts list, we should add them if we don't (perhaps tell them directly if we support them)

	//if we want to auto vote, then we'll vote for them!
	if (con.autoVoteWith != nil &&
		*con.autoVoteWith == proposed.InitialVote.Address()) ||
		(con.autoVoteForNode != nil &&
			*con.autoVoteForNode == accounts.AccountFromPubBytes(proposed.NodeID).Address) {
		//we have a reason to auto vote for this node
		err := con.VoteFor(proposed.NodeID, con.autoStakeAmount)
		return err
	}
	return nil
}

// used to vote for a candidate (normally ourselves)
func (con *ConsensusNode) VoteFor(candidateNodeID crypto.PublicKey, stakeAmount *big.Int) error {
	if con.thisCandidate != nil && bytes.Equal(con.thisCandidate.NodeID, candidateNodeID) {
		return fmt.Errorf("cannot vote for self")
	}
	if _, exists := con.candidates[string(candidateNodeID)]; !exists {
		return fmt.Errorf("candidate doesn't exist, we might want to save it locally in the future and see if we get the candidate later")
	}
	vote := utils.NewVote(con.spendingKey.PublicKey, stakeAmount)
	err := vote.SignTo(*con.candidates[string(candidateNodeID)], con.spendingKey)
	if err != nil {
		return err
	}
	con.votesSeen[string(candidateNodeID)] = append(con.votesSeen[string(candidateNodeID)], &vote)
	con.candidateStakeValues[string(candidateNodeID)].Add(con.candidateStakeValues[string(candidateNodeID)], stakeAmount)

	return con.netLogic.Propagate(vote)
}

// propose this node as a witness for the network.
func (con *ConsensusNode) ProposeCandidacy() error {
	log.Println("proposing self for candidacy")
	if con.thisCandidate == nil {
		con.thisCandidate = con.generateCandidacy()
	}
	//TODO: update candidacy for this rounds

	con.thisCandidate.InitialVote = utils.NewVote(con.spendingKey.PublicKey, big.NewInt(1)) //TODO: change this to a real staking amount
	con.thisCandidate.Round = con.currentRound + 1
	con.thisCandidate.InitialVote.SignTo(*con.thisCandidate, con.spendingKey)

	con.netLogic.Propagate(*con.thisCandidate)

	return nil
}

// generate a mostly blank self candidate proposal
func (con *ConsensusNode) generateCandidacy() *utils.Candidate {
	foo := con.handlingType
	thisCandidate := utils.Candidate{
		Round:         0, //TODO: have the round numbers
		NodeID:        con.participation.PublicKey,
		ConsensusPool: int8(foo),
		VRFKey:        con.vrfKey.PublicKey,
		NetworkString: con.netLogic.GetConnectionString(),
	}
	return &thisCandidate
}
