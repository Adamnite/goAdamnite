package statedb

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/log15"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	// emptyRoot is the known root hash of an empty trie.
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

// StateDB structs within the adamnite protocol are used to store anything within the merkle trie.
// StateDB take care of caching and storing nested state.
// It's the general query interface to retrieve
// - Witnesses
// - Accounts
type StateDB struct {
	db           Database
	originalRoot common.Hash // The pre-state root, before any changes were made
	trie         Trie
	hasher       crypto.KeccakState

	snapAccounts  map[common.Hash][]byte
	snapWitnesses map[common.Hash][]byte

	journal *journal

	stateObjects        map[common.Address]*stateObject
	stateObjectsDirty   map[common.Address]struct{}
	stateObjectsPending map[common.Address]struct{}
	lock                sync.Mutex
	dbErr               error
}

func New(root common.Hash, db Database) (*StateDB, error) {
	tr, err := db.OpenTrie(root)
	if err != nil {
		return nil, err
	}

	stateDB := &StateDB{
		db:                  db,
		originalRoot:        root,
		trie:                tr,
		snapAccounts:        make(map[common.Hash][]byte),
		snapWitnesses:       make(map[common.Hash][]byte),
		stateObjects:        make(map[common.Address]*stateObject),
		stateObjectsDirty:   make(map[common.Address]struct{}),
		stateObjectsPending: make(map[common.Address]struct{}),
		journal:             newJournal(),
	}

	return stateDB, nil
}

func (s *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObj := s.GetOrNewStateObj(addr)
	if stateObj != nil {
		stateObj.SetNonce(nonce)
	}
}

func (s *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObj := s.GetOrNewStateObj(addr)
	if stateObj != nil {
		stateObj.AddBalance(amount)
	}
}

func (s *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	stateObj := s.GetOrNewStateObj(addr)
	if stateObj != nil {
		stateObj.SetBalance(amount)
	}
}

func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	s.Finalise(deleteEmptyObjects)

	// ToDO: implement prefetcher logic

	for addr := range s.stateObjectsPending {
		if obj := s.stateObjects[addr]; !obj.deleted {
			obj.updateRoot(s.db)
		}
	}

	// ToDO: implement prefetcher logic

	usedAddrs := make([][]byte, 0, len(s.stateObjectsPending))
	for addr := range s.stateObjectsPending {
		if obj := s.stateObjects[addr]; obj.deleted {
			s.deleteStateObject(obj)
		} else {
			s.updateStateObject(obj)
		}
		usedAddrs = append(usedAddrs, common.CopyBytes(addr[:])) // Copy needed for closure
	}

	// ToDO: implement prefetcher logic

	if len(s.stateObjectsPending) > 0 {
		s.stateObjectsPending = make(map[common.Address]struct{})
	}

	return s.trie.Hash()
}

func (s *StateDB) Finalise(deleteEmptyObjects bool) {
	addressesToPrefetch := make([][]byte, 0, len(s.journal.dirties))
	for addr := range s.journal.dirties {
		obj, exist := s.stateObjects[addr]
		if !exist {
			continue
		}
		if obj.suicided || (deleteEmptyObjects && obj.empty()) {
			obj.deleted = true
		} else {
			obj.finalise(true)
		}
		s.stateObjectsDirty[addr] = struct{}{}
		s.stateObjectsPending[addr] = struct{}{}

		addressesToPrefetch = append(addressesToPrefetch, common.CopyBytes(addr[:]))
	}

	// ToDO: implement prefetcher logic
}

// Commit writes the state to the underlying in-memory trie database.
func (s *StateDB) Commit(deleteEmptyObjects bool) (common.Hash, error) {
	if s.dbErr != nil {
		return common.Hash{}, fmt.Errorf("commit aborted due to earlier error: %v", s.dbErr)
	}
	// Finalize any pending changes and merge everything into the tries
	s.IntermediateRoot(deleteEmptyObjects)

	// Commit objects to the trie, measuring the elapsed time
	codeWriter := s.db.TrieDB().DiskDB().NewBatch()
	for addr := range s.stateObjectsDirty {
		if obj := s.stateObjects[addr]; !obj.deleted {
			// Write any storage changes in the state object to its storage trie
			if err := obj.CommitTrie(s.db); err != nil {
				return common.Hash{}, err
			}
		}
	}
	if len(s.stateObjectsDirty) > 0 {
		s.stateObjectsDirty = make(map[common.Address]struct{})
	}
	if codeWriter.ValueSize() > 0 {
		if err := codeWriter.Write(); err != nil {
			log15.Crit("Failed to commit dirty codes", "error", err)
		}
	}

	// The onleaf func is called _serially_, so we can reuse the same account
	// for unmarshalling every time.
	var account Account
	root, err := s.trie.Commit(func(_ [][]byte, _ []byte, leaf []byte, parent common.Hash) error {
		if err := msgpack.Unmarshal(leaf, &account); err != nil {
			return nil
		}
		if account.Root != emptyRoot {
			s.db.TrieDB().Reference(account.Root, parent)
		}
		return nil
	})

	// ToDO: implement snapshot
	return root, err
}

// Database retrieves the low level database supporting the lower level trie ops.
func (s *StateDB) Database() Database {
	return s.db
}

func (s *StateDB) GetOrNewStateObj(addr common.Address) *stateObject {
	stateObj := s.getStateObject(addr)
	if stateObj == nil {
		stateObj, _ = s.createStateObject(addr)
	}
	return stateObj
}

func (s *StateDB) createStateObject(addr common.Address) (newObj, prev *stateObject) {
	prev = s.getDeletedStateObject(addr)

	newObj = newObject(s, addr, Account{})
	newObj.setNonce(0)

	if prev == nil {
		s.journal.append(createObjectChange{account: &addr})
	} else {
		s.journal.append(resetObjectChange{prev: prev, prevdestruct: false})
	}

	s.setStateObject(newObj)

	if prev != nil && !prev.deleted {
		return newObj, prev
	}
	return newObj, nil
}

func (s *StateDB) setStateObject(obj *stateObject) {
	s.stateObjects[obj.Address()] = obj
}

func (s *StateDB) getStateObject(addr common.Address) *stateObject {
	if obj := s.getDeletedStateObject(addr); obj != nil && !obj.deleted {
		return obj
	}
	return nil
}

func (s *StateDB) getDeletedStateObject(addr common.Address) *stateObject {
	if obj := s.stateObjects[addr]; obj != nil {
		return obj
	}

	var (
		data *Account
		err  error
	)

	enc, err := s.trie.TryGet(addr.Bytes())
	if err != nil {
		s.setError(fmt.Errorf("getDeletedStateObject (%x) error: %v", addr.Bytes(), err))
		return nil
	}
	if len(enc) == 0 {
		return nil
	}

	data = new(Account)
	if err := msgpack.Unmarshal(enc, data); err != nil {
		log15.Error("Failed to decode state object", "addr", addr, "err", err)
		return nil
	}

	obj := newObject(s, addr, *data)
	s.setStateObject(obj)
	return obj
}

// deleteStateObject removes the given object from the state trie.
func (s *StateDB) deleteStateObject(obj *stateObject) {
	// Delete the account from the trie
	addr := obj.Address()
	if err := s.trie.TryDelete(addr[:]); err != nil {
		s.setError(fmt.Errorf("deleteStateObject (%x) error: %v", addr[:], err))
	}
}

// updateStateObject writes the given object to the trie.
func (s *StateDB) updateStateObject(obj *stateObject) {
	// Encode the account and update the account trie
	addr := obj.Address()

	data, err := msgpack.Marshal(obj)
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	}
	if err = s.trie.TryUpdate(addr[:], data); err != nil {
		s.setError(fmt.Errorf("updateStateObject (%x) error: %v", addr[:], err))
	}

	// ToDO: implement snapshot
}

func (s *StateDB) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}
