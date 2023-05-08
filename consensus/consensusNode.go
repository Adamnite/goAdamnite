package consensus

import (
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
)

// The node that can handle consensus systems.
// Is built on top of a networking node, using a netNode to handle the networking interactions
type ConsensusNode struct {
	netLogic     networking.NetNode
	handlingType consensusHandlingTypes
	//TODO: VRF keys, aka, the keyset used for applying for votes and whatnot. should have its own type upon implementation here
	spendingKey crypto.PrivateKey // each consensus node is forced to have its own account to spend from.
	chain       *core.Blockchain  //we need to keep the chain
	ocdbLink    string            //off chain database, if running the VM verification, this should be local.
}

func NewAConsensus() ConsensusNode {
	return ConsensusNode{
		netLogic:     *networking.NewNetNode(common.Address{}),
		handlingType: SecondaryTransactions,
	}
}
func NewConsensus() ConsensusNode {
	return ConsensusNode{
		netLogic: *networking.NewNetNode(common.Address{}),
	}
}
