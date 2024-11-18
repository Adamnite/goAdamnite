package core

import (
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
)

type NewBlockEvent struct {
	Block *types.Block
}

type ImportBlockEvent struct {
	Block *types.Block
}

type NewTxsEvent struct{ Txs []*types.Transaction }

type ChainEvent struct {
	Block *types.Block
	Hash  common.Hash
}

type ChainSideEvent struct {
	Block *types.Block
}

type ChainHeadEvent struct{ Block *types.Block }
