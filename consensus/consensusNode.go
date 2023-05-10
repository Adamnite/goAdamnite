package consensus

import (
	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
)

// The node that can handle consensus systems.
// Is built on top of a networking node, using a netNode to handle the networking interactions
type ConsensusNode struct {
	netLogic     *networking.NetNode
	handlingType consensusHandlingTypes
	//TODO: VRF keys, aka, the keyset used for applying for votes and whatnot. should have its own type upon implementation here
	spendingKey crypto.PrivateKey      // each consensus node is forced to have its own account to spend from.
	chain       *blockchain.Blockchain //we need to keep the chain
	ocdbLink    string                 //off chain database, if running the VM verification, this should be local.
	vm          *VM.Machine
}

func NewAConsensus() (ConsensusNode, error) {
	conNode, err := newConsensus(networking.NewNetNode(common.Address{}))
	conNode.handlingType = SecondaryTransactions
	return conNode, err
}

func newConsensus(hostingNode *networking.NetNode) (ConsensusNode, error) {
	return ConsensusNode{
		netLogic:     hostingNode,
		handlingType: NetworkingOnly,
	}, nil
}
