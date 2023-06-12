package consensus

import (
	"crypto/rand"
	"errors"
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
	thisCandidateA *utils.Candidate
	thisCandidateB *utils.Candidate
	poolsA         *Witness_pool //store all the vote data for chamber a elections
	poolsB         *Witness_pool //store all the vote data for chamber b elections

	netLogic     *networking.NetNode
	handlingType networking.NetworkTopLayerType

	spendingAccount accounts.Account // each consensus node is forced to have its own account to spend from.
	participation   accounts.Account
	vrfKey          crypto.PrivateKey
	state           *statedb.StateDB
	chain           *blockchain.Blockchain //we need to keep the chain
	ocdbLink        string                 //off chain database, if running the VM verification, this should be local.
	vm              *VM.Machine

	autoVoteForNode *crypto.PublicKey //the NodeID that is running the node
	autoVoteWith    *common.Address
	autoStakeAmount *big.Int

	untrustworthyWitnesses map[string]uint64 //nodeID -> keep track of how many times witness was marked as untrustworthy
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
		untrustworthyWitnesses: make(map[string]uint64),
	}
	vrfKey, _ := crypto.GenerateVRFKey(rand.Reader)
	con.vrfKey = vrfKey
	if err := hostingNode.AddFullServer(state, chain, con.ReviewTransaction, con.ReviewCandidacy, con.ReviewVote); err != nil {
		log.Printf("error:%v", err)
		return nil, err
	}
	return &con, nil
}
func (con ConsensusNode) CanReviewType(t networking.NetworkTopLayerType) bool {
	return t.IsTypeIn(con.handlingType)
}
func (con ConsensusNode) CanReview(t int8) bool {
	return networking.NetworkTopLayerType(t).IsTypeIn(con.handlingType)
}
func (con ConsensusNode) IsActiveWitnessFor(processType networking.NetworkTopLayerType) bool {
	if !con.CanReviewType(processType) {
		//we aren't handling that processes type
		return false
	}
	switch processType { //see what type of transaction it is
	case networking.PrimaryTransactions:
		return con.poolsA.IsActiveWitness((*crypto.PublicKey)(&con.spendingAccount.PublicKey))
	case networking.SecondaryTransactions:
		return con.poolsB.IsActiveWitness((*crypto.PublicKey)(&con.spendingAccount.PublicKey))
	}
	return false
}

func (con *ConsensusNode) ReviewTransaction(transaction *utils.Transaction) error {
	//give this a quick look over, review it, if its good, add it locally and propagate it out, otherwise, ignore it.
	var tReviewType networking.NetworkTopLayerType
	if transaction.VMInteractions != nil {
		//the call the VM
		tReviewType = networking.SecondaryTransactions
	} else {
		tReviewType = networking.PrimaryTransactions
	}

	if !con.IsActiveWitnessFor(tReviewType) {
		return nil //we aren't a witness, so no reason to review this transaction.
	}
	//TODO: review the transaction under the appropriate node method

	//this is the global consensus review. Even if we aren't a witness, this is called anytime we see a transaction go past.
	//an error will prevent the transaction from being propagated past us
	return nil
}

func (con *ConsensusNode) ReviewBlock(block *utils.Block) error {
	//TODO: give this a quick look over, review it, if its good, add it locally and propagate it out, otherwise, ignore it.
	//this is the global consensus review. Even if we aren't a witness, this is called anytime we see a block go past.
	//an error will prevent the transaction from being propagated past us
	if !con.CanReview(block.Header.TransactionType) {
		return nil //we aren't fit to review this at all. We therefor cant say if its good or bad, so we just share it
	}
	if block.Header.TransactionType == int8(networking.PrimaryTransactions) {
		//TODO: also record this for our own records!
		valid, err := con.ValidateChamberABlock(block)
		if !valid {
			//TODO: this can be a false error if it's just because the chain isnt set (but then we shouldnt be reviewing...)
			return err
		} else {
			return nil
		}

	}
	return nil
}

// validates if a header is truthful and can be traced back to the genesis block
func (n *ConsensusNode) ValidateHeader(header *utils.BlockHeader) error {
	if header == nil {
		return errors.New("unknown block")
	}

	if header.Number.Cmp(big.NewInt(0)) == 0 {
		// our header comes from genesis block
		return nil
	}

	parentHeader := ConvertBlockHeader(n.chain.GetHeader(header.ParentBlockID, big.NewInt(0).Sub(header.Number, big.NewInt(1))))
	if parentHeader == nil || parentHeader.Number.Cmp(big.NewInt(0).Sub(header.Number, big.NewInt(1))) != 0 || parentHeader.Hash() != header.ParentBlockID {
		return errors.New("unknown parent block")
	}

	if parentHeader.Timestamp.Add(maxTimePerRound).Before(header.Timestamp) {
		return errors.New("invalid timestamp")
	}

	return nil
}
