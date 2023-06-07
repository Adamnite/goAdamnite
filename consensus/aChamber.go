package consensus

import (
	"errors"
	"math/big"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

func NewAConsensus(account accounts.Account) (*ConsensusNode, error) {
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
	n.poolsA, err = newWitnessPool(0, networking.PrimaryTransactions, []byte{})
	//TODO: the genesis round is 0, with seed {}, we should get the current round and seed info from nodes we know
	return
}

func (n *ConsensusNode) isANode() bool {
	return networking.PrimaryTransactions.IsTypeIn(n.handlingType)
}

func (n *ConsensusNode) ValidateHeader(header *BlockHeader, interval int64) error {
	if header == nil {
		return errors.New("unknown block")
	}

	if header.Number.Cmp(big.NewInt(0)) == 0 {
		// our header comes from genesis block
		return nil
	}

	parentHeader := ConvertBlockHeader(n.chain.GetHeader(header.ParentBlockID, big.NewInt(0).Sub(header.Number, big.NewInt(1))))
	if parentHeader == nil || parentHeader.Number.Cmp(big.NewInt(0).Sub(header.Number, big.NewInt(1))) != 0 || parentHeader.Hash() != header.ParentBlockID {
		return errors.New("unknown parent block")
	}

	if parentHeader.Timestamp+interval > header.Timestamp {
		return errors.New("invalid timestamp")
	}

	return nil
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
