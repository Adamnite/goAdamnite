package consensus

import (
	"github.com/adamnite/go-adamnite/VM"
<<<<<<< Updated upstream
=======
	"github.com/adamnite/go-adamnite/utils/bytes"
	"github.com/adamnite/go-adamnite/utils/math"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/consensus/pendingHandling"
>>>>>>> Stashed changes
	"github.com/adamnite/go-adamnite/networking"
)

// for methods that only apply to the B chamber members
func NewBConsensus(codeServer string) (*ConsensusNode, error) {
	conNode, err := newConsensus(nil, nil)
	conNode.handlingType = networking.SecondaryTransactions
	conNode.ocdbLink = codeServer

	return conNode, err
}

func (bNode *ConsensusNode) isBNode() bool {
	return bNode.handlingType == networking.SecondaryTransactions
}

// run the claimed changes again to verify that we have the same results.
func (bNode *ConsensusNode) VerifyRun(runtimeClaim VM.RuntimeChanges) (bool, *VM.RuntimeChanges, error) {
	if !bNode.isBNode() {
		return false, nil, ErrNotBNode
	}
	ourChanges := runtimeClaim.CleanCopy()
	err := bNode.ProcessRun(ourChanges)
	trustable := runtimeClaim.Equal(ourChanges)
	return trustable, ourChanges, err
}

// take an incomplete runtime claim, and handle the process. The answer is assigned to the claim.
func (bNode *ConsensusNode) ProcessRun(claim *VM.RuntimeChanges) error {
	if !bNode.isBNode() {
		return ErrNotBNode
	}
	if bNode.vm == nil {
		vm, err := VM.NewVirtualMachineWithContract(bNode.ocdbLink, nil)
		if err != nil {
			return err
		}
		bNode.vm = vm
	}
	newClaim, err := bNode.vm.CallWith(bNode.ocdbLink, claim)
	if err != nil {
		return err
	}
<<<<<<< Updated upstream
	*claim = *newClaim
	return err
=======
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
	vmb := utils.NewWorkingVMBlock(
		bNode.chain.CurrentHeader().Hash(),
		bNode.spendingAccount.PublicKey,
		bytes.Hash{},
		bytes.Hash{},
		bytes.Hash{},
		big.NewInt(0), //TODO: last block, +1
		0,             //TODO: check if this would be the end of a round. If so, change it.
		[]utils.TransactionType{},
	)
	//now we need to run all of those transactions. While leaving the state handler open for more applications
	transactions, err := conState.RunOnUntil(bNode.state, maxTransactionsPerBlock, nil, vmb) //TODO: add the new round channel
	if err != nil {
		return err
	}
	//TODO: here we should actually use the transactions
	log.Println(transactions)
	log.Println("transactions were returned!")

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
>>>>>>> Stashed changes
}
