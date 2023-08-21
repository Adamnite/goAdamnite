package blockchain

import (
	"math/big"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/databaseDeprecated/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/params"
)

type StateProcessor struct {
	config             *params.ChainConfig // Chain configuration options
	bc                 *Blockchain         // Canonical block chain
	engine             dpos.AdamniteDPOS   // Consensus engine used for block rewards
	vmInstances        []VM.Machine
	localDBAPIEndpoint string
}

// NewStateProcessor initializes a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *Blockchain, engine dpos.AdamniteDPOS) *StateProcessor {
	return &StateProcessor{
		config:             config,
		bc:                 bc,
		engine:             engine,
		vmInstances:        []VM.Machine{},
		localDBAPIEndpoint: "http://127.0.0.1:5001/",
	}
}

func (p *StateProcessor) Process(block *types.Block, statedb *statedb.StateDB, cfg VM.VMConfig, gasPrice *big.Int) (uint64, error) {
	var (
		usedGas = new(uint64)
		header  = block.Header()
	)
	// Mutate the block and state according to any hard-fork specs
	// Iterate over and process the individual transactions
	if cfg.CodeGetter == nil {
		getObject := VM.NewAPICodeGetter(p.localDBAPIEndpoint)
		cfg.CodeGetter = getObject.GetCode
	}
	for i, tx := range block.Body().Transactions {
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		v, err := ApplyTransaction(
			p.config,
			p.bc,
			nil,
			gasPrice,
			statedb,
			header,
			tx,
			usedGas,
			cfg,
			VM.NewBlockContext( //TODO: someone with a better understanding of block structure should review this!
				header.DBWitness, //coinbase Address //TODO:SOMEONE REVIEW THIS!
				tx.ATEMax(),
				p.bc.CurrentBlock().Number(),
				big.NewInt(block.ReceivedAt.UnixMicro()),
				big.NewInt(1),
				tx.Cost()))
		if err != nil {
			return 0, err
		}
		p.vmInstances = append(p.vmInstances, *v)
	}
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Body().Transactions)
	if p.localDBAPIEndpoint != "" { //just check that this server is in fact, running a DB
		for _, v := range p.vmInstances {
			err := v.UploadMachinesContract(p.localDBAPIEndpoint)
			if err != nil {
				return 0, err
			}
		}
	}
	return *usedGas, nil
}

func ApplyTransaction(config *params.ChainConfig, bc *Blockchain, author *common.Address, gp *big.Int,
	statedb *statedb.StateDB, header *types.BlockHeader, tx *types.Transaction, usedGas *uint64, vmcfg VM.VMConfig, blockContext VM.BlockContext) (*VM.Machine, error) {

	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, err
	}

	vmenv := VM.NewVM(
		statedb, //stateDB
		&vmcfg,  //vm config
		config)  //chain config
	// Apply the transaction to the current state (included in the env)
	_, gas, _, err := core.ApplyMessage(vmenv, msg)
	if err != nil {
		return nil, err
	}

	*usedGas += gas

	return vmenv, err
}
