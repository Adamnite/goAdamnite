package dpos

import (
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/params"
)

type ChainHeaderReader interface {
	// CurrentHeader retrieves the current header of the chain.
	CurrentHeader() *types.BlockHeader

	// GetHeaderByNumber retrieves the header from the database by number.
	GetHeaderByNumber(number uint64) *types.BlockHeader

	// GetHeaderByHash retrieves the header from the database by hash.
	GetHeaderByHash(hash common.Hash) *types.BlockHeader
}

type ChainReader interface {
	// Config retrieves the blockchain's configuration.
	Config() *params.ChainConfig

	ChainHeaderReader

	// GetBlockByHash retrieves the block from the database by hash.
	GetBlockByHash(hash common.Hash) *types.Block

	// GetBlockByNumber retrieves the block from the database by number.
	GetBlockByNumber(number uint64) *types.Block
}

// Engine is an algorithm agnostic consensus engine.
type Engine interface {
	// Witness retrives the Adamnite address that generated the given block
	Witness(header *types.BlockHeader) (common.Address, error)

	// VerifyHeader checks whether a header conforms to the consensus rules of the given engine.
	VerifyHeader(header *types.BlockHeader, chain ChainHeaderReader) error

	// Close terminates all background threads maintained by the engine.
	Close() error
}

type DPOS interface {
	Engine

	// GetRoundNumber retrieves the number of current round.
	GetRoundNumber() uint64
}
