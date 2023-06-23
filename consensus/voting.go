package consensus

// a file to clean up the voting process from the consensus node file.
import (
	"bytes"
	"fmt"
	"log"
	"math/big"

	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/safe"
)

// used to vote for a candidate (normally ourselves)
func (con *ConsensusNode) VoteFor(candidate *utils.Candidate, stakeAmount *big.Int) error {
	vote := utils.NewVote(con.spendingAccount.PublicKey, stakeAmount)
	err := vote.SignTo(*candidate, *con.spendingAccount)
	if err != nil {
		return err
	}

	if networking.NetworkTopLayerType(candidate.ConsensusPool).IsTypeIn(con.handlingType) {
		//check that we are running the network type of this candidate, and if so, store the candidate if we haven't already
		var pool *Witness_pool
		switch candidate.ConsensusPool {
		case uint8(networking.PrimaryTransactions):
			pool = con.poolsA
		case uint8(networking.SecondaryTransactions):
			pool = con.poolsB
		}
		can := pool.GetCandidate(candidate.GetWitnessPub())
		if can == nil {
			//check if we have that candidate saved, if not, add it!
			if err := pool.AddCandidate(candidate); err != nil {
				return err
			}
		}
	}

	err = con.ReviewVote(vote)
	if err != nil {
		return err
	}
	return con.netLogic.Propagate(vote)
}

// review authenticity of a vote, as well as recording it for our own records. If no errors are returned, will propagate further
func (con *ConsensusNode) ReviewVote(vote utils.Voter) error {
	var pool *Witness_pool
	if !networking.NetworkTopLayerType(vote.PoolCategory).IsTypeIn(con.handlingType) {
		//we aren't setup to handle this type
		return fmt.Errorf("this consensus node does not have the ability to verify this vote")
	}
	if uint8(networking.PrimaryTransactions) == (vote.PoolCategory) {
		//the vote is for pool A and we can handle it!
		pool = con.poolsA
	} else if uint8(networking.SecondaryTransactions) == (vote.PoolCategory) {
		pool = con.poolsB

	}
	candidate := pool.GetCandidate((*crypto.PublicKey)(&vote.To))
	if candidate == nil {
		return fmt.Errorf("we don't have that account saved, we could throw an error, or check for that candidate first, maybe save unknown votes to check again at the end?")
	}

	//TODO: check the balance of these voters as well as verify the vote!
	verified := candidate.VerifyVote(vote)
	if !verified {
		return ErrVoteUnVerified
	}

	//assuming by here it is legit.
	return pool.AddVote(candidate.Round, &vote)
}

// review a candidate proposal. The Consensus node may add a vote on. If no errors are returned, assume it is fine to forward along.
func (con *ConsensusNode) ReviewCandidacy(proposed utils.Candidate) error {

	//review that the initial vote is signed correctly
	if !proposed.VerifyVote(proposed.InitialVote) {
		// log.Println("someone lied in a vote")
		//TODO: assume malicious attempt and distrust this witness
		return ErrVoteUnVerified
	}
	if networking.PrimaryTransactions.IsIn(proposed.ConsensusPool) && con.poolsA != nil {
		if err := con.poolsA.AddCandidate(&proposed); err != nil {
			return err
		}
	} else if networking.SecondaryTransactions.IsIn(proposed.ConsensusPool) && con.poolsB != nil {
		if err := con.poolsB.AddCandidate(&proposed); err != nil {
			return err
		}
	}

	//if we made it here, this is most likely a viable candidate
	//TODO: check if we have this candidate in our networking contacts list, we should add them if we don't (perhaps tell them directly if we support them)

	//if we want to auto vote, then we'll vote for them!
	if (con.autoVoteWith != nil &&
		*con.autoVoteWith == proposed.InitialVote.Address()) ||
		(con.autoVoteForNode != nil &&
			bytes.Equal(*con.autoVoteForNode, proposed.NodeID)) {
		//we have a reason to auto vote for this node
		return con.VoteFor(&proposed, con.autoStakeAmount)
	}
	return nil
}

// propose this node as a witness for the network types listed. candidacyTypes should be passed as the mask of types you are applying for.
// if you wish to apply to all types you are handling, pass 0
func (con *ConsensusNode) ProposeCandidacy(candidacyTypes uint8) error {
	log.Println("proposing self for candidacy")
	if err := con.updateAllOurCandidates(); err != nil {
		return err
	}
	if candidacyTypes == 0 {
		candidacyTypes = uint8(con.handlingType)
	}
	if candidacyTypes == uint8(networking.NetworkingOnly) {
		//assume they are intentionally doing this to prevent further applications
		if len(con.poolsA.newRoundStartedCaller) != 0 {
			con.poolsA.newRoundStartedCaller = []func(){con.continuosHandler}
		}
		//TODO: check that poolsB, if it's running, we need to prevent further applications to it as well.
		return nil
	}
	if networking.PrimaryTransactions.IsIn(candidacyTypes) { //we're proposing ourselves for chamber A
		addLocalCandidate := func() {
			pool := con.poolsA
			newCon, err := safe.GetItem[*utils.Candidate](con.thisCandidateA).UpdatedCandidate(uint64(pool.currentWorkingRoundID.Get()+1), pool.GetCurrentSeed(), con.vrfKey, uint64(pool.GetApplyingRound().GetStartTime().Unix()), *con.spendingAccount)
			if err != nil {
				log.Printf("error updating candidate for round %v. Err: %v", pool.currentWorkingRoundID, err)
				return
			}
			con.thisCandidateA.Set(newCon)
			if err := con.poolsA.AddCandidate(safe.GetItem[*utils.Candidate](con.thisCandidateA)); err != nil {
				log.Printf("error producing newer candidate for round %v. Err: %v", pool.currentWorkingRoundID, err)
				panic(err)
			}

			if err := con.netLogic.Propagate(safe.GetItem[*utils.Candidate](con.thisCandidateA)); err != nil {
				panic(err)
			}
		}
		addLocalCandidate()
		con.poolsA.AddNewRoundCaller(addLocalCandidate)
		con.poolsA.AddNewRoundCaller(con.continuosHandler)
	}
	if networking.SecondaryTransactions.IsIn(candidacyTypes) { //we're proposing ourselves for chamber B
		newBRoundActions := func() {
			if err := con.poolsB.AddCandidate(con.thisCandidateB.Get().(*utils.Candidate)); err != nil {
				panic(err)
			}
			if err := con.netLogic.Propagate(con.thisCandidateB.Get().(*utils.Candidate)); err != nil {
				panic(err)
			}
		}
		newBRoundActions()
		con.poolsB.AddNewRoundCaller(newBRoundActions)
	}
	return nil
}

func (con *ConsensusNode) updateAllOurCandidates() (err error) {
	if con.poolsA != nil {
		if con.thisCandidateA == nil {
			con.thisCandidateA = safe.NewSafeItem(con.generateCandidacy())
			safe.GetItem[*utils.Candidate](con.thisCandidateA).ConsensusPool = uint8(networking.PrimaryTransactions)
		}
		updated, err := con.getUpdatedCandidacy(safe.GetItem[*utils.Candidate](con.thisCandidateA), con.poolsA)
		if err != nil {
			return err
		}
		con.thisCandidateA.Set(updated)
	}
	if con.poolsB != nil {
		if con.thisCandidateB == nil {
			con.thisCandidateB = safe.NewSafeItem(con.generateCandidacy())
			safe.GetItem[*utils.Candidate](con.thisCandidateB).ConsensusPool = uint8(networking.SecondaryTransactions)
		}
		updated, err := con.getUpdatedCandidacy(safe.GetItem[*utils.Candidate](con.thisCandidateB), con.poolsA)
		if err != nil {
			return err
		}
		con.thisCandidateB.Set(updated)
	}
	return nil
}

// generate a mostly blank self candidate proposal
func (con *ConsensusNode) generateCandidacy() *utils.Candidate {
	foo := con.handlingType
	if con.autoStakeAmount == nil {
		con.autoStakeAmount = big.NewInt(0)
	}
	thisCandidate, _ := utils.NewCandidate(0, []byte{}, con.vrfKey, 0, uint8(foo), con.netLogic.GetOwnContact().ConnectionString, con.nodeAccount.PublicKey, *con.spendingAccount, con.autoStakeAmount)
	return thisCandidate
}

// create an updated version of the candidacy provided
func (con *ConsensusNode) getUpdatedCandidacy(candidacy *utils.Candidate, pool *Witness_pool) (*utils.Candidate, error) {
	//TODO: get the round start time! right now it's set to 0
	return candidacy.UpdatedCandidate(uint64(pool.currentWorkingRoundID.Get()), pool.GetCurrentSeed(), con.vrfKey, 0, *con.spendingAccount)
}
