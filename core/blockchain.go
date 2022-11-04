package core

import (
	"sync"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/event"
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
	blocks        []types.Block // memory cache
	accountStates map[common.Address]accountSet

	// events
	importBlockFeed event.Feed
	scope           event.SubscriptionScope

	chainlock sync.RWMutex
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

func (bc *Blockchain) Config() *params.ChainConfig { return bc.chainConfig }

func (bc *Blockchain) CurrentHeader() *types.BlockHeader {
	return bc.CurrentBlock().Header()
}

func (bc *Blockchain) GetHeader(hash common.Hash, number uint64) *types.BlockHeader {
	return nil
}

func (bc *Blockchain) GetHeaderByHash(hash common.Hash) *types.BlockHeader {
	return nil
}

func (bc *Blockchain) GetHeaderByNumber(number uint64) *types.BlockHeader {
	return bc.blocks[number].Header()
}

func (bc *Blockchain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return nil
}

func (bc *Blockchain) GetBlockByHash(hash common.Hash) *types.Block {
	return nil
}

func (bc *Blockchain) GetBlockByNumber(number uint64) *types.Block {
	return nil
}
func (bc *Blockchain) StateAt(root common.Hash) (*statedb.StateDB, error) {
	return nil, nil
}

func (bc *Blockchain) CurrentBlock() *types.Block {
	return &bc.blocks[len(bc.blocks)-1]
}

func (bc *Blockchain) WriteBlock(block *types.Block) error {
	bc.chainlock.Lock()
	defer bc.chainlock.Unlock()

	bc.blocks = append(bc.blocks, *block)
	return nil
}

func (bc *Blockchain) AddImportedBlock(block *types.Block) error {
	bc.chainlock.Lock()
	defer bc.chainlock.Unlock()

	currentBlock := bc.CurrentBlock()

	if currentBlock.Numberu64() >= block.Numberu64() {
		return nil
	}

	bc.blocks = append(bc.blocks, *block)

	bc.importBlockFeed.Send(ImportBlockEvent{Block: block})
	return nil
}

func (bc *Blockchain) SubscribeImportBlockEvent(ch chan<- ImportBlockEvent) event.Subscription {
	return bc.scope.Track(bc.importBlockFeed.Subscribe(ch))
}
