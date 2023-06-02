package consensus

import (
	"errors"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

func NewAConsensus(account accounts.Account) (*ConsensusNode, error) {
	//TODO: setup the chain data and whatnot
	n, err := newConsensus(nil, nil)
	n.spendingAccount = account
	n.handlingType = networking.PrimaryTransactions
	return n, err
}

func (n *ConsensusNode) isANode() bool {
	return n.handlingType == networking.PrimaryTransactions
}

func (n *ConsensusNode) ValidateBlock(block *Block) (bool, error) {
	if !n.isANode() {
		return false, ErrNotANode
	}

	tmp := block
	// iterate until the genesis block
	for (tmp.Header.ParentBlockID != common.Hash{}) {
		if n.chain == nil {
			return false, errors.New("chain not set")
		}

		parentBlock := n.chain.GetBlockByHash(tmp.Header.ParentBlockID)
		if parentBlock == nil {
			// parent block does not exist on chain
			// thus, proposed block is not valid so we mark the witness as untrustworthy
			n.untrustworthyWitnesses[block.Header.Witness] += 1
			return false, nil
		}

		// note: temporary adapter until we start using consensus structures across the rest of codebase
		tmp = ConvertBlock(parentBlock)
	}
	return true, nil
}
