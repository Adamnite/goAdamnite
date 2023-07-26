package consensus

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/consensus/pendingHandling"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

func NewAConsensus(account *accounts.Account) (*ConsensusNode, error) {
	//TODO: setup the chain data and whatnot
	n, err := newConsensus(nil, nil)
	if err != nil {
		return nil, err
	}
	n.spendingAccount = account

	return n, n.AddAConsensus()
}
func (n *ConsensusNode) AddAConsensus() (err error) {
	//adds primary transactions handling type
	n.handlingType = n.handlingType ^ networking.PrimaryTransactions
	n.poolsA, err = NewWitnessPool(0, networking.PrimaryTransactions, []byte{})
	//TODO: the genesis round is 0, with seed {}, we should get the current round and seed info from nodes we know
	return
}

func (n *ConsensusNode) isANode() bool {
	return networking.PrimaryTransactions.IsTypeIn(n.handlingType)
}

// handles the full automation of validating a block and logging results to the local witness states storage.
func (n *ConsensusNode) ValidateChamberABlock(block *utils.Block) (bool, error) {
	valid, err := n.getAChamberBlockValidity(block)
	if err != nil {
		return valid, err
	}
	//the block is good, check that the transactions we have saved as pending aren't in here (and if they are, remove them)
	//however, this cleanup isn't necessary for if this is valid, so we'll do it in its own thread
	pendingHandling.RemoveAllFrom(block.Transactions, n.transactionQueue)
	return valid, n.poolsA.ActiveWitnessReviewed(&block.Header.Witness, valid, block.Header.Number.Uint64())
}

// purely reviews the blocks validity, but does not change any states based on the results
func (aCon ConsensusNode) getAChamberBlockValidity(block *utils.Block) (bool, error) {
	if !aCon.isANode() {
		return false, ErrNotANode
	}
	tmp := block
	//TODO: this can be removed and replaced by the looping solution once chain storage is correct.
	if !tmp.VerifySignature() {
		log.Printf("signature invalid on block with hash %x", tmp.Hash().Bytes())
		return false, errors.New("block most likely has been tampered with and the signature cannot be verified")
	}

	// iterate until the genesis block
	for (tmp.Header.ParentBlockID != common.Hash{}) {
		if aCon.chain == nil {
			return false, errors.New("chain not set")
		}

		parentBlock := aCon.chain.GetBlockByHash(tmp.Header.ParentBlockID)
		if parentBlock == nil {
			// parent block does not exist on chain
			// thus, proposed block is not valid so we mark the witness as untrustworthy
			aCon.untrustworthyWitnesses[string(block.Header.Witness)] += 1
			return false, fmt.Errorf("error block invalid")
			// aCon.poolsA.ActiveWitnessReviewed(&block.Header.Witness, false, block.Header.Number.Uint64())
		}
		//TODO: uncomment that once we have the chain blocks corrected
		// if !tmp.VerifySignature() {
		// 	log.Printf("signature invalid on block with hash %x", tmp.Hash().Bytes())
		// 	return false, errors.New("block most likely has been tampered with and the signature cannot be verified")
		// }
		// note: temporary adapter until we start using consensus structures across the rest of codebase
		tmp = ConvertBlock(parentBlock)
	}
	return true, nil
}

// creates the continuous thread method that takes all the queued transactions, checks they can be run, and add its to a "working" block
func (aCon *ConsensusNode) continuosHandler() { //TODO: rename this
	//please only call on nodes setup to handle a consensus...
	go func() {
		//TODO: for the love of god clean this up
		aCon.transactionQueue.SortQueue()

		for aCon.poolsA.IsActiveWitness((*crypto.PublicKey)(&aCon.spendingAccount.PublicKey)) {
			//we want to keep this going while we aren't the leader...
			if aCon.IsActiveWitnessLeadFor(networking.PrimaryTransactions) {
				aCon.actAsLead()
			}
		}
	}()
}

// trying to breakup this massive go routine
func (aCon *ConsensusNode) actAsLead() {
	transactions := []*utils.BaseTransaction{}
	//TODO: get the trie base point
	stateLocking := sync.RWMutex{}
	for aCon.IsActiveWitnessLeadFor(networking.PrimaryTransactions) {
		//TODO: replace this with an "is active witness leader"
		possibleT := aCon.transactionQueue.Pop()
		if possibleT == nil {
			//we're out of transactions and need to wait for a new transaction to be sent
			continue
		}
		//do the things like check the time, signature, whatnot
		stateLocking.RLock()
		if ok, err := aCon.IsStillApplicable(possibleT); !ok {
			log.Printf("error with that transaction, %v", err)
			stateLocking.RUnlock()
			continue
		}
		stateLocking.RUnlock()
		//t can on occasion actually be a VM transaction instead of one intended for us, so just put it back and start again
		if possibleT.GetType() != utils.Transaction_Basic {
			aCon.transactionQueue.AddIgnoringPast(possibleT)
			continue
		}
		t := possibleT.(*utils.BaseTransaction)
		stateLocking.Lock()
		aCon.state.SubBalance(t.FromAddress(), t.Amount)
		aCon.state.AddBalance(t.To, t.Amount)
		stateLocking.Unlock()
		//TODO: charge them some gas for sending this

		transactions = append(transactions, t)
		if len(transactions) >= maxTransactionsPerBlock {
			parent := aCon.chain.CurrentHeader()

			workingBlock := utils.NewBlock(
				parent.Hash(),
				aCon.spendingAccount.PublicKey,
				common.Hash{}, //TODO: get these values
				common.Hash{},
				common.Hash{},
				big.NewInt(0).Add(parent.Number, big.NewInt(1)),
				transactions,
			)

			workingBlock.Header.Round = uint64(aCon.poolsA.currentWorkingRoundID.Get())
			if err := workingBlock.Sign(*aCon.spendingAccount); err != nil {
				log.Printf("error signing block %v", err)
			}
			// workingBlock
			transactions = []*utils.BaseTransaction{}
			log.Printf("submitting block with hash %x", workingBlock.Hash().Bytes())
			if ok, err := aCon.getAChamberBlockValidity(workingBlock); ok {
				// aCon.chain.WriteBlock()//TODO: we should like, be able to save blocks...
				aCon.netLogic.Propagate(workingBlock)
			} else {
				//TODO: we should try and see what went wrong and see if we can fix it
				log.Println("self made block could not pe properly validated with reason")
				log.Println(err)
			}
		}
	}
}
