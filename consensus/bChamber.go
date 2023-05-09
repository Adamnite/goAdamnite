package consensus

import (
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/VM"
	"github.com/adamnite/go-adamnite/networking"
)

// for methods that only apply to the B chamber members
func NewBConsensus() (ConsensusNode, error) {
	conNode, err := newConsensus(networking.NewNetNode(common.Address{0}))
	conNode.handlingType = PrimaryTransactions

	return conNode, err
}

// run the claimed changes again to verify that we have the same results.
func (bNode *ConsensusNode) VerifyRun(runtimeClaim VM.RuntimeChanges) (bool, error) {
	if bNode.vm == nil {
		vm, err := VM.NewVirtualMachineWithContract(bNode.ocdbLink, nil)
		if err != nil {
			return false, err
		}
		bNode.vm = vm
	}
	ourChanges, err := bNode.vm.CallWith(bNode.ocdbLink, runtimeClaim.CleanCopy())

	return runtimeClaim.Equal(*ourChanges), err
}
