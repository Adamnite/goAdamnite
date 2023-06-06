package consensus

import (
	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/networking"
)

// for methods that only apply to the B chamber members
func NewBConsensus(codeServer string) (*ConsensusNode, error) {
	conNode, err := newConsensus(nil, nil)
	if err != nil {
		return conNode, err
	}
	return conNode, conNode.AddBConsensus(codeServer)
}

// for adding support for B chamber
func (bNode *ConsensusNode) AddBConsensus(codeServer string) (err error) {
	bNode.handlingType = bNode.handlingType ^ networking.SecondaryTransactions
	bNode.ocdbLink = codeServer
	bNode.poolsB, err = newWitnessPool(0, networking.SecondaryTransactions, []byte{})
	//TODO: the genesis round is 0, with seed {}, we should get the current round and seed info from nodes we know
	return err
}

func (bNode *ConsensusNode) isBNode() bool {
	return networking.SecondaryTransactions.IsTypeIn(bNode.handlingType)
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
	*claim = *newClaim
	return err
}
