package consensus

import (
	"bytes"
	"crypto/rand"
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
	thisCandidate *utils.Candidate
	poolsA        *witness_pool
	poolsB        *witness_pool
	votesSeen     map[string][]*utils.Voter   //candidate nodeID->voters array
	candidates    map[string]*utils.Candidate //candidate nodeID->Candidate(just sorting by nodeID)

	netLogic     *networking.NetNode
	handlingType networking.NetworkTopLayerType

	spendingAccount accounts.Account // each consensus node is forced to have its own account to spend from.
	participation   accounts.Account
	vrfKey          crypto.PrivateKey
	state           *statedb.StateDB
	chain           *blockchain.Blockchain //we need to keep the chain
	ocdbLink        string                 //off chain database, if running the VM verification, this should be local.
	vm              *VM.Machine

	autoVoteForNode *common.Address
	autoVoteWith    *common.Address
	autoStakeAmount *big.Int

	untrustworthyWitnesses map[common.Address]uint64 //keep track of how many times witness was marked as untrustworthy
}

func newConsensus(state *statedb.StateDB, chain *blockchain.Blockchain) (*ConsensusNode, error) {
	participation, err := accounts.GenerateAccount()
	if err != nil {
		return nil, err
	}
	hostingNode := networking.NewNetNode(participation.Address)
	con := ConsensusNode{
		netLogic:               hostingNode,
		handlingType:           networking.NetworkingOnly,
		state:                  state,
		chain:                  chain,
		participation:          *participation,
		votesSeen:              make(map[string][]*utils.Voter),
		candidates:             make(map[string]*utils.Candidate),
		untrustworthyWitnesses: make(map[common.Address]uint64),
	}
	vrfKey, _ := crypto.GenerateVRFKey(rand.Reader)
	con.vrfKey = vrfKey
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
	return nil
}

// review a candidate proposal. The Consensus node may add a vote on. If no errors are returned, assume it is fine to forward along.
func (con *ConsensusNode) ReviewCandidacy(proposed utils.Candidate) error {
	log.Println("reviewing candidate!")

	//review that the initial vote is signed correctly
	if !proposed.VerifyVote(proposed.InitialVote) {
		log.Println("someone lied in a vote")
		return ErrVoteUnVerified
	}

	if proposed.ConsensusPool == int8(networking.PrimaryTransactions) && con.poolsA != nil {
		if err := con.poolsA.AddCandidate(&proposed); err != nil {
			return err
		}
	} else if proposed.ConsensusPool == int8(networking.SecondaryTransactions) && con.poolsB != nil {
		if err := con.poolsB.AddCandidate(&proposed); err != nil {
			return err
		}
	}

	con.candidates[string(proposed.NodeID)] = &proposed
	con.votesSeen[string(proposed.NodeID)] = []*utils.Voter{&proposed.InitialVote}

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
	vote := utils.NewVote(con.spendingAccount.PublicKey, stakeAmount)
	err := vote.SignTo(*con.candidates[string(candidateNodeID)], con.spendingAccount)

	if err != nil {
		return err
	}

	con.poolsA.AddVote(con.candidates[string(candidateNodeID)].Round, &candidateNodeID, &vote)
	con.votesSeen[string(candidateNodeID)] = append(con.votesSeen[string(candidateNodeID)], &vote)

	return con.netLogic.Propagate(vote)
}

// propose this node as a witness for the network.
func (con *ConsensusNode) ProposeCandidacy() error {
	log.Println("proposing self for candidacy")
	if con.thisCandidate == nil {
		con.thisCandidate = con.generateCandidacy()
	}
	if con.poolsA != nil {
		//TODO: get the round start times!
		con.thisCandidate, _ = con.thisCandidate.UpdatedCandidate(con.poolsA.currentRound+1, con.poolsA.getNextSeed(), con.vrfKey, 0, con.spendingAccount)
	}

	//TODO: update candidacy for this rounds

	con.thisCandidate.InitialVote = utils.NewVote(con.spendingAccount.PublicKey, big.NewInt(1)) //TODO: change this to a real staking amount
	// con.thisCandidate.Round = con.currentRound + 1
	con.thisCandidate.InitialVote.SignTo(*con.thisCandidate, con.spendingAccount)

	con.netLogic.Propagate(*con.thisCandidate)

	return nil
}

// generate a mostly blank self candidate proposal
func (con *ConsensusNode) generateCandidacy() *utils.Candidate {
	foo := con.handlingType
	if con.autoStakeAmount == nil {
		con.autoStakeAmount = big.NewInt(0)
	}
	thisCandidate, _ := utils.NewCandidate(0, []byte{}, con.vrfKey, 0, uint8(foo), con.netLogic.GetOwnContact().ConnectionString, con.participation.PublicKey, con.spendingAccount, con.autoStakeAmount)
	return thisCandidate
}
