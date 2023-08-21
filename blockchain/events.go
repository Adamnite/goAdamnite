package blockchain

import (
	"github.com/adamnite/go-adamnite/core/types"
)

type ImportBlockEvent struct {
	Block *types.Block
}

type ChainSideEvent struct {
	Block *types.Block
}

type ChainHeadEvent struct{ Block *types.Block }
