package adm

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/core/vm"
	"github.com/adamnite/go-adamnite/log15"
)

func (adm *AdamniteImpl) stateAtBlock(block *types.Block, reexec uint64, base *statedb.StateDB, checkLive bool) (statedb *statedb.StateDB, err error) {
	var (
		current  *types.Block
		database statedb.Database
		report   = true
		origin   = block.Numberu64()
	)
	// Check the live database first if we have the state fully available, use that.
	if checkLive {
		statedb, err = adm.blockchain.StateAt(block.Header().StateRoot)
		if err == nil {
			return statedb, nil
		}
	}
	if base != nil {
		// The optional base statedb is given, mark the start point as parent block
		statedb, database, report = base, base.Database(), false
		current = adm.blockchain.GetBlock(block.Hash(), block.Numberu64()-1)
	} else {
		// Otherwise try to reexec blocks until we find a state or reach our limit
		current = block

		// Create an ephemeral trie.Database for isolating the live one. Otherwise
		// the internal junks created by tracing will be persisted into the disk.
		database = statedb.NewDatabaseWithConfig(adm.chainDB, &trie.Config{Cache: 16})

		// If we didn't check the dirty database, do check the clean one, otherwise
		// we would rewind past a persisted block (specific corner case is chain
		// tracing from the genesis).
		if !checkLive {
			statedb, err = statedb.New(current.Header().StateRoot, database, nil)
			if err == nil {
				return statedb, nil
			}
		}
		// Database does not have the state for the given block, try to regenerate
		for i := uint64(0); i < reexec; i++ {
			if current.Numberu64() == 0 {
				return nil, errors.New("genesis state is missing")
			}
			parent := adm.blockchain.GetBlock(current.Header().ParentHash, current.Numberu64()-1)
			if parent == nil {
				return nil, fmt.Errorf("missing block %v %d", current.Header().ParentHash, current.Numberu64()-1)
			}
			current = parent

			statedb, err = statedb.New(current.Header().StateRoot, database, nil)
			if err == nil {
				break
			}
		}
		if err != nil {
			switch err.(type) {
			case *trie.MissingNodeError:
				return nil, fmt.Errorf("required historical state unavailable (reexec=%d)", reexec)
			default:
				return nil, err
			}
		}
	}
	// State was available at historical point, regenerate
	var (
		start  = time.Now()
		logged time.Time
		parent common.Hash
	)
	for current.Numberu64() < origin {
		// Print progress logs if long enough time elapsed
		if time.Since(logged) > 8*time.Second && report {
			log15.Info("Regenerating historical state", "block", current.Numberu64()+1, "target", origin, "remaining", origin-current.Numberu64()-1, "elapsed", time.Since(start))
			logged = time.Now()
		}
		// Retrieve the next block to regenerate and process it
		next := current.Numberu64() + 1
		if current = adm.blockchain.GetBlockByNumber(next); current == nil {
			return nil, fmt.Errorf("block #%d not found", next)
		}

		// Finalize the state so any modifications are written to the trie
		root, err := statedb.Commit(true)
		if err != nil {
			return nil, err
		}
		statedb, err = statedb.New(root, database, nil)
		if err != nil {
			return nil, fmt.Errorf("state reset after block %d failed: %v", current.Numberu64(), err)
		}
		database.TrieDB().Reference(root, common.Hash{})
		if parent != (common.Hash{}) {
			database.TrieDB().Dereference(parent)
		}
		parent = root
	}
	if report {
		nodes, imgs := database.TrieDB().Size()
		log15.Info("Historical state regenerated", "block", current.Numberu64(), "elapsed", time.Since(start), "nodes", nodes, "preimages", imgs)
	}
	return statedb, nil
}

// stateAtTransaction returns the execution environment of a certain transaction.
func (adm *AdamniteImpl) stateAtTransaction(block *types.Block, txIndex int, reexec uint64) (core.Message, vm.BlockContext, *statedb.StateDB, error) {
	// Short circuit if it's genesis block.
	if block.Numberu64() == 0 {
		return nil, vm.BlockContext{}, nil, errors.New("no transaction in genesis")
	}
	// Create the parent state database
	parent := adm.blockchain.GetBlock(block.Header().ParentHash, block.Numberu64()-1)
	if parent == nil {
		return nil, vm.BlockContext{}, nil, fmt.Errorf("parent %#x not found", block.Header().ParentHash)
	}
	// Lookup the statedb of parent block from the live database,
	// otherwise regenerate it on the flight.
	statedb, err := adm.stateAtBlock(parent, reexec, nil, true)
	if err != nil {
		return nil, vm.BlockContext{}, nil, err
	}
	if txIndex == 0 && len(block.Body().Transactions) == 0 {
		return nil, vm.BlockContext{}, statedb, nil
	}
	// Recompute transactions up to the target index.
	signer := types.MakeSigner(adm.blockchain.Config(), block.Number())
	for idx, tx := range block.Body().Transactions {
		// Assemble the transaction call message and return if the requested offset
		msg, _ := tx.AsMessage(signer)
		txContext := vm.NewTxContext(msg)
		context := vm.NewBlockContext(*tx.To(), 100, block.Number(), new(big.Int).SetUint64(block.Header().Time), tx.Amount(), tx.ATEPrice())
		if idx == txIndex {
			return msg, context, statedb, nil
		}
		// Not yet the searched for transaction, execute on top of the current state
		vmenv := vm.NewVM(context, txContext, statedb, adm.blockchain.Config(), vm.VMConfig{})
		statedb.Prepare(tx.Hash(), block.Hash(), idx)
		if _, _, _, err := core.ApplyMessage(vmenv, msg); err != nil {
			return nil, vm.BlockContext{}, nil, fmt.Errorf("transaction %#x failed: %v", tx.Hash(), err)
		}
		// Ensure any modifications are committed to the state
		// Only delete empty objects if EIP158/161 (a.k.a Spurious Dragon) is in effect
		statedb.Finalise(true)
	}
	return nil, vm.BlockContext{}, nil, fmt.Errorf("transaction index %d out of range for block %#x", txIndex, block.Hash())
}
