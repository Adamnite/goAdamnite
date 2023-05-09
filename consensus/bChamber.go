package consensus

import (
	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/networking"
)

// for methods that only apply to the B chamber members
func NewBConsensus(codeServer string) (ConsensusNode, error) {
	conNode, err := newConsensus(networking.NewNetNode(common.Address{0}))
	conNode.handlingType = SecondaryTransactions
	conNode.ocdbLink = codeServer

	return conNode, err
}

func (bNode *ConsensusNode) isBNode() bool {
	return bNode.handlingType == SecondaryTransactions
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
