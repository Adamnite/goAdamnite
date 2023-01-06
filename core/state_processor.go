package core

import (
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/params"
)

type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *Blockchain         // Canonical block chain
	engine dpos.AdamniteDPOS   // Consensus engine used for block rewards
}

// NewStateProcessor initializes a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *Blockchain, engine dpos.AdamniteDPOS) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

func (p *StateProcessor) Process(block *types.Block, statedb *statedb.StateDB, cfg vm.Config) (uint64, error) {
	var (
		usedGas = new(uint64)
		header  = block.Header()
	)
	// Mutate the block and state according to any hard-fork specs
	// Iterate over and process the individual transactions
	for i, tx := range block.Body().Transactions {
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil {
			return 0, err
		}

	}
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Body().Transactions)

	return *usedGas, nil
}

func ApplyTransaction(config *params.ChainConfig, bc *Blockchain, author *common.Address, gp *GasPool, statedb *statedb.StateDB, header *types.BlockHeader, tx *types.Transaction, usedGas *uint64, cfg vm.Config) error {
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return err
	}

	context := NewVMContext(msg, header, bc, author)
	vmenv := vm.NewVM(context, statedb, config, cfg)
	// Apply the transaction to the current state (included in the env)
	_, gas, _, err := ApplyMessage(vmenv, msg, gp)
	if err != nil {
		return nil, err
	}

	*usedGas += gas

	return err
}
