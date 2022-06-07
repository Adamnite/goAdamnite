package core

import (
	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/params"
)

const (
	EpochDuration = 27 * 6
)

type Blockchain struct {
	genesisBlock *types.Block

	chainConfig *params.ChainConfig

	db adamnitedb.Database

	engine  dpos.DPOS
	witness types.Witness

	// For demo version
	blocks []types.Block // memory cache
}

func NewBlockchain(db adamnitedb.Database, chainConfig *params.ChainConfig, engine dpos.DPOS) (*Blockchain, error) {
	bc := &Blockchain{
		chainConfig: chainConfig,
		db:          db,
		engine:      engine,
	}

	// demo logic
	genesis := DefaultTestnetGenesisBlock()
	block, err := genesis.Write(db)
	if err != nil {
		return nil, err
	}
	bc.blocks = append(bc.blocks, *block)
	bc.genesisBlock = block

	return bc, nil
}

func (bc *Blockchain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return nil
}

func (bc *Blockchain) StateAt(root common.Hash) (*statedb.StateDB, error) {
	return nil, nil
}

func (bc *Blockchain) CurrentBlock() *types.Block {
	return &bc.blocks[len(bc.blocks)-1]
}
