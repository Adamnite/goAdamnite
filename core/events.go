package core

import "github.com/adamnite/go-adamnite/core/types"

type NewBlockEvent struct {
	Block *types.Block
}

type ImportBlockEvent struct {
	Block *types.Block
}
