package core

import (
	"math/big"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/core/vm"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/params"
)

type StateProcessor struct {
	config      *params.ChainConfig // Chain configuration options
	bc          *Blockchain         // Canonical block chain
	engine      dpos.AdamniteDPOS   // Consensus engine used for block rewards
	vmInstances []vm.Machine
}

// NewStateProcessor initializes a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *Blockchain, engine dpos.AdamniteDPOS) *StateProcessor {
	return &StateProcessor{
		config:      config,
		bc:          bc,
		engine:      engine,
		vmInstances: []vm.Machine{},
	}
}

func (p *StateProcessor) Process(block *types.Block, statedb *statedb.StateDB, cfg vm.VMConfig, gasPrice *big.Int) (uint64, error) {
	var (
		usedGas = new(uint64)
		header  = block.Header()
	)
	// Mutate the block and state according to any hard-fork specs
	// Iterate over and process the individual transactions
	for i, tx := range block.Body().Transactions {
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		//TODO: actually get the blockContext...
		v, err := ApplyTransaction(p.config, p.bc, nil, gasPrice, statedb, header, tx, usedGas, cfg, vm.BlockContext{})
		if err != nil {
			return 0, err
		}
		p.vmInstances = append(p.vmInstances, *v)
	}
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Body().Transactions)
	for i, v := range p.vmInstances {

	}
	return *usedGas, nil
}

func ApplyTransaction(config *params.ChainConfig, bc *Blockchain, author *common.Address, gp *big.Int,
	statedb *statedb.StateDB, header *types.BlockHeader, tx *types.Transaction, usedGas *uint64, vmcfg vm.VMConfig, blockContext vm.BlockContext) (*vm.Machine, error) {

	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, err
	}

	vmenv := vm.NewVM(
		statedb,      //stateDB
		blockContext, //blockContext
		vm.TxContext{ //transaction context
			Origin:   msg.From(),
			GasPrice: gp},
		&vmcfg, //vm config
		config) //chain config
	// Apply the transaction to the current state (included in the env)
	_, gas, _, err := ApplyMessage(vmenv, msg)
	if err != nil {
		return nil, err
	}

	*usedGas += gas

	return vmenv, err
}
