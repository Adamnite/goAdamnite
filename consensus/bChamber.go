package consensus

import (
	"fmt"
	"log"
	"math/big"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/consensus/pendingHandling"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
)

// for methods that only apply to the B chamber members
func NewBConsensus(codeServer VM.DBCacheAble) (*ConsensusNode, error) {
	conNode, err := newConsensus(nil, nil)
	if err != nil {
		return conNode, err
	}
	return conNode, conNode.AddBConsensus(codeServer)
}

// for adding support for B chamber
func (bNode *ConsensusNode) AddBConsensus(codeServer VM.DBCacheAble) (err error) {
	//TODO: verify that there is a server running at that endpoint, and we can in fact, access it
	bNode.handlingType = bNode.handlingType ^ networking.SecondaryTransactions
	bNode.ocdbLink = codeServer
	bNode.poolsB, err = NewWitnessPool(0, networking.SecondaryTransactions, []byte{})
	//TODO: the genesis round is 0, with seed {}, we should get the current round and seed info from nodes we know
	return err
}

func (bNode *ConsensusNode) isBNode() bool {
	return networking.SecondaryTransactions.IsTypeIn(bNode.handlingType)
}

// run the claimed changes again to verify that we have the same results.
func (bNode *ConsensusNode) VerifyRun(runtimeClaim utils.RuntimeChanges) (bool, *utils.RuntimeChanges, error) {
	if !bNode.isBNode() {
		return false, nil, ErrNotBNode
	}
	ourChanges := runtimeClaim.CleanCopy()
	err := bNode.ProcessRun(ourChanges)
	trustable := runtimeClaim.Equal(ourChanges)
	return trustable, ourChanges, err
}

// take an incomplete runtime claim, and handle the process. The answer is assigned to the claim.
func (bNode *ConsensusNode) ProcessRun(claim *utils.RuntimeChanges) error {
	if !bNode.isBNode() {
		return ErrNotBNode
	}
	//TODO: delete
	return nil
}

// verifies a block, and optionally saves the results to the OCDB
func (bNode *ConsensusNode) VerifyVMBlock(block *utils.VMBlock, saveToDB bool) (bool, error) {
	conStates, err := pendingHandling.NewContractStateHolder(bNode.ocdbLink)
	if err != nil {
		return false, err
	}
	for _, t := range block.Transactions {
		//first we need to get all of these transactions loaded into a contractState handler.
		switch t.GetType() {
		case utils.Transaction_Basic:
			return false, fmt.Errorf("wrong transaction type found in block") //TODO: make real error
		case utils.Transaction_VM_Call:
			if err := conStates.QueueTransaction(t.(*utils.VMCallTransaction)); err != nil {
				return false, err
			}
			//TODO: add a case to handle new contract creation
		}
	}
	state := bNode.state
	if !saveToDB {
		//if we aren't saving locally, then we don't want to apply the transactions either
		state = bNode.state.Copy()
	}
	if err := conStates.RunAll(state); err != nil {
		return false, err
	}
	//TODO: call the conStates to get the final hashes and make sure they're the same
	return true, nil
}
func (bNode *ConsensusNode) StartVMContinuosHandling() error {
	//TODO: listed below
	//only run this when we're a witness
	//wait until we're selected as the lead.
	//when we're the lead, get all of our transactions in order
	//once we have all our initial transactions in order, we filter out expired ones, and add the valid ones to a contractState handler
	//

	//start of us being lead
	bNode.transactionQueue.SortQueue()
	conState, err := pendingHandling.NewContractStateHolder(bNode.ocdbLink)
	if err != nil {
		return err
	}
	//handle the backlog
	var firstPutBack utils.TransactionType = nil //to check if we are putting the same node at the back, that way we don't get into a continuos loop
	t := bNode.transactionQueue.Pop()
	for (firstPutBack == nil || !firstPutBack.Equal(t)) && t != nil {
		if t == nil {
			//we're out of transactions and need to wait for a new transaction to be sent
			continue
		}
		//TODO: filter out expired contracts
		if t.GetType() == utils.Transaction_Basic {
			//this is a chamber A transaction, so lets put it to the back of the queue
			bNode.transactionQueue.AddIgnoringPast(t)
			if firstPutBack == nil {
				firstPutBack = t
			}
			continue
		}
		//by here we assume it's valid
		conState.QueueTransaction(t)

		//add this for the next time around the loop
		t = bNode.transactionQueue.Pop()
	}
	doneProcessing := make(chan (any))
	newTransactionAdded := bNode.transactionQueue.GetOnUpdate()
	go func() {
		for {
			select {
			case <-doneProcessing:
				return
			case newIndex := <-newTransactionAdded:
				newItem := bNode.transactionQueue.Get(newIndex)
				if newItem != nil && newItem.GetType() != utils.Transaction_Basic {
					//this is a new VM calling transaction
					bNode.transactionQueue.Remove(newItem)
					if err := conState.QueueTransaction(newItem); err != nil {
						log.Printf("error adding new transaction to the queue: %v", err)
					}
				}
			}
		}
	}()
	//now we need to run all of those transactions. While leaving the state handler open for more applications
	transactions, err := conState.RunOnUntil(bNode.state, maxTransactionsPerBlock, nil)
	if err != nil {
		return err
	}
	//TODO: here we should actually use the transactions
	log.Println(transactions)
	log.Println("transactions were returned!")
	vmb := utils.NewWorkingVMBlock(
		bNode.chain.CurrentHeader().Hash(),
		bNode.spendingAccount.PublicKey,
		common.Hash{},
		common.Hash{},
		common.Hash{},
		big.NewInt(0), //TODO: last block, +1
		0,             //TODO: check if this would be the end of a round. If so, change it.
		transactions,
	)
	err = vmb.Sign(*bNode.spendingAccount)
	if err != nil {
		log.Println("error with the signing: ", err)
	}
	if err := bNode.netLogic.Propagate(vmb); err != nil {
		log.Println(err)
	}
	bNode.poolsB.ActiveWitnessReviewed(
		&vmb.Header.Witness, //us, but with different steps
		true,
		vmb.Header.Number.Uint64(),
	)
	return nil
}
