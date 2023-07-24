package statedb

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/common"
)

func TestTrieUpdate(t *testing.T) {
	db := rawdb.NewMemoryDB()
	state, _ := New(bytes.Hash{}, NewDatabase(db))

	for i := byte(0); i < 255; i++ {
		addr := common.BytesToAddress([]byte{i})
		state.AddBalance(addr, big.NewInt(int64(i)))
		state.SetNonce(addr, uint64(i))
	}

	rootHash := state.IntermediateRoot(false)
	state.Database().TrieDB().Commit(rootHash, false, nil)
}

func TestTrieStateUpdate(t *testing.T) {
	transDb := rawdb.NewMemoryDB()
	finalDb := rawdb.NewMemoryDB()
	transState, _ := New(bytes.Hash{}, NewDatabase(transDb))
	finalState, _ := New(bytes.Hash{}, NewDatabase(finalDb))

	modifyAccount := func(state *StateDB, addr bytes.Address, i, tweak byte) {
		state.SetBalance(addr, big.NewInt(int64(i*10)+int64(tweak)))
		state.SetNonce(addr, uint64(10*i+tweak))
	}

	for i := byte(0); i < 255; i++ {
		modifyAccount(transState, bytes.Address{i}, i, 0)
	}

	transState.IntermediateRoot(false)

	for i := byte(0); i < 255; i++ {
		modifyAccount(transState, bytes.Address{i}, i, 1)
		modifyAccount(finalState, bytes.Address{i}, i, 1)
	}

	transRoot, err := transState.Commit(false)
	if err != nil {
		t.Fatalf("failed to commit transition state: %v", err)
	}
	if err = transState.Database().TrieDB().Commit(transRoot, false, nil); err != nil {
		t.Errorf("can not commit trie %v to persistent database", transRoot.Hex())
	}

	finalRoot, err := finalState.Commit(false)
	if err != nil {
		t.Fatalf("failed to commit final state: %v", err)
	}
	if err = finalState.Database().TrieDB().Commit(finalRoot, false, nil); err != nil {
		t.Errorf("can not commit trie %v to persistent database", finalRoot.Hex())
	}

	it := finalDb.NewIterator(nil, nil)
	for it.Next() {
		key, fvalue := it.Key(), it.Value()
		tvalue, err := transDb.Get(key)
		if err != nil {
			t.Errorf("entry missing from the transition database: %x -> %x", key, fvalue)
		}
		if !bytes.Equal(fvalue, tvalue) {
			t.Errorf("the value associate key %x is mismatch,: %x in transition database ,%x in final database", key, tvalue, fvalue)
		}
	}
	it.Release()

	it = transDb.NewIterator(nil, nil)
	for it.Next() {
		key, tvalue := it.Key(), it.Value()
		fvalue, err := finalDb.Get(key)
		if err != nil {
			t.Errorf("extra entry in the transition database: %x -> %x", key, it.Value())
		}
		if !bytes.Equal(fvalue, tvalue) {
			t.Errorf("the value associate key %x is mismatch,: %x in transition database ,%x in final database", key, tvalue, fvalue)
		}
	}
}
