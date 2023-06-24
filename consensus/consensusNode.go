package consensus

import (
	"crypto/rand"
	"errors"
	"log"
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/adm/database"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/consensus/pendingHandling"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/rpc"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

// The node that can handle consensus systems.
// Is built on top of a networking node, using a netNode to handle the networking interactions
type ConsensusNode struct {
	thisCandidateA   *utils.Candidate
	thisCandidateB   *utils.Candidate
	poolsA           *Witness_pool //store all the vote data for chamber a elections
	poolsB           *Witness_pool //store all the vote data for chamber b elections
	transactionQueue *pendingHandling.TransactionQueue
	netLogic         *networking.NetNode
	handlingType     networking.NetworkTopLayerType

	spendingAccount accounts.Account // each consensus node is forced to have its own account to spend from.
	participation   accounts.Account
	vrfKey          crypto.PrivateKey
	state           *database.StateDatabase
	chain           *blockchain.Blockchain //we need to keep the chain
	ocdbLink        string                 //off chain database, if running the VM verification, this should be local.
	vm              *VM.Machine

	autoVoteForNode *crypto.PublicKey //the NodeID that is running the node
	autoVoteWith    *common.Address
	autoStakeAmount *big.Int

	untrustworthyWitnesses map[string]uint64 //nodeID -> keep track of how many times witness was marked as untrustworthy
}

func newConsensus(state *database.StateDatabase, chain *blockchain.Blockchain) (*ConsensusNode, error) {
	participation, err := accounts.GenerateAccount()
	if err != nil {
		return nil, err
	}
	hostingNode := networking.NewNetNode(participation.Address)
	con := ConsensusNode{
		transactionQueue:       pendingHandling.NewQueue(true),
		netLogic:               hostingNode,
		handlingType:           networking.NetworkingOnly,
		state:                  state,
		chain:                  chain,
		participation:          *participation,
		untrustworthyWitnesses: make(map[string]uint64),
	}
	vrfKey, _ := crypto.GenerateVRFKey(rand.Reader)
	con.vrfKey = vrfKey
	if err := hostingNode.AddFullServer(
		state,
		chain,
		con.ReviewTransaction,
		con.ReviewBlock,
		con.ReviewCandidacy,
		con.ReviewVote,
	); err != nil {
		log.Printf("error:%v", err)
		return nil, err
	}
	return &con, nil
}
func (con *ConsensusNode) Close() {
	//close all mappings
	go func() {
		for k := range con.votesSeen {
			delete(con.votesSeen, k)
		}
	}()
	go func() {
		for k := range con.candidateStakeValues {
			delete(con.candidateStakeValues, k)
		}
	}()
	go func() {
		for k := range con.candidates {
			delete(con.candidates, k)
		}
	}()
	con.netLogic.Close()
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
	//TODO: sending nite and calling a VM should check that they are calling to the same target. And will need further input on it's processing
	if !con.CanReviewType(tReviewType) {
		//see if we're able to handle this type of transaction
		return nil
	}
	//check if the transaction is expired (signature verification should be done when balance is checked)
	if transaction.Time.Add(maxTimePerRound.Duration()).Before(time.Now().UTC()) {
		//the transaction expired
		return rpc.ErrBadForward
	}

	con.transactionQueue.AddToQueue(transaction)
	return nil
	//TODO: delete the rest of my notes to self
	//actual way
	//witness's all receive the transactions, each witness has a turn in the witness order, and through that turn, takes the transaction
	//this is the global consensus review. Even if we aren't a witness, this is called anytime we see a transaction go past.
	//an error will prevent the transaction from being propagated past us
}

// called when a new block is received over the network. We review it, and only return an error if we aren't setup to handle it
func (con *ConsensusNode) ReviewBlock(block utils.Block) error {
	//this is the global consensus review. Even if we aren't a witness, this is called anytime we see a block go past.
	//an error will prevent the transaction from being propagated past us
	if !con.CanReview(block.Header.TransactionType) {
		return nil //we aren't fit to review this at all. We therefor cant say if its good or bad, so we just share it
	}
	if block.Header.TransactionType == int8(networking.PrimaryTransactions) {
		//TODO: also record this for our own records!
		valid, err := con.ValidateChamberABlock(&block)
		if !valid {
			//TODO: this can be a false error if it's just because the chain isn't set (but then we shouldn't be reviewing...)
			// return err
			log.Println("error with block validity")
			log.Println(err)
			//TODO: no, we want to throw this error!!!
			return nil
		}

	}
	//TODO: if we are chamber a, or b, either way we should log the transactions to our state
	return nil
}

// validates if a header is truthful and can be traced back to the genesis block
func (n *ConsensusNode) ValidateHeader(header *utils.BlockHeader) (bool, error) {
	if header == nil {
		return false, errors.New("unknown block")
	}

	if header.Number.Cmp(big.NewInt(0)) == 0 {
		// our header comes from genesis block
		return true, nil
	}

	parentHeader := ConvertBlockHeader(n.chain.GetHeader(header.ParentBlockID, big.NewInt(0).Sub(header.Number, big.NewInt(1))))
	if parentHeader == nil || parentHeader.Number.Cmp(big.NewInt(0).Sub(header.Number, big.NewInt(1))) != 0 || parentHeader.Hash() != header.ParentBlockID {
		return false, errors.New("unknown parent block")
	}

	if parentHeader.Timestamp.Add(maxTimePerRound.Duration()).Before(header.Timestamp) {
		return false, errors.New("invalid timestamp")
	}

	return true, nil
}

func (con *ConsensusNode) ConnectTo(endpoint string) error {
	return con.netLogic.ConnectToSeed(endpoint)
}
