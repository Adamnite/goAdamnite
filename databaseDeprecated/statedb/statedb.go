package statedb

import (
	"fmt"
	"hash"
	"math/big"
	"sort"
	"sync"

	"github.com/adamnite/go-adamnite/common"

	encoding "github.com/vmihailenco/msgpack/v5"
	log "github.com/sirupsen/logrus"
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
type revision struct {
	id           int
	journalIndex int
}

type StateDB struct {
	db           Database
	originalRoot common.Hash // The pre-state root, before any changes were made
	trie         Trie
	hasher       hash.Hash

	snapAccounts  map[common.Hash][]byte
	snapWitnesses map[common.Hash][]byte

	journal        *journal
	validRevisions []revision
	nextRevisionId int

	stateObjects        map[common.Address]*stateObject
	stateObjectsDirty   map[common.Address]struct{}
	stateObjectsPending map[common.Address]struct{}
	lock                sync.Mutex
	dbErr               error

	thash, bhash common.Hash
	txIndex      int
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

func (s *StateDB) Copy() *StateDB {
	// Copy all the basic fields, initialize the memory ones
	state := &StateDB{
		db:                  s.db,
		trie:                s.db.CopyTrie(s.trie),
		stateObjects:        make(map[common.Address]*stateObject, len(s.journal.dirties)),
		stateObjectsPending: make(map[common.Address]struct{}, len(s.stateObjectsPending)),
		stateObjectsDirty:   make(map[common.Address]struct{}, len(s.journal.dirties)),
		snapAccounts:        make(map[common.Hash][]byte),
		snapWitnesses:       make(map[common.Hash][]byte),
		journal:             newJournal(),
	}
	// Copy the dirty states, logs, and preimages
	for addr := range s.journal.dirties {
		// As documented [here](https://github.com/ethereum/go-ethereum/pull/16485#issuecomment-380438527),
		// and in the Finalise-method, there is a case where an object is in the journal but not
		// in the stateObjects: OOG after touch on ripeMD prior to Byzantium. Thus, we need to check for
		// nil
		if object, exist := s.stateObjects[addr]; exist {
			// Even though the original object is dirty, we are not copying the journal,
			// so we need to make sure that anyside effect the journal would have caused
			// during a commit (or similar op) is already applied to the copy.
			state.stateObjects[addr] = object.deepCopy(state)

			state.stateObjectsDirty[addr] = struct{}{}   // Mark the copy dirty to force internal (code/state) commits
			state.stateObjectsPending[addr] = struct{}{} // Mark the copy pending to force external (account) commits
		}
	}
	// Above, we don't copy the actual journal. This means that if the copy is copied, the
	// loop above will be a no-op, since the copy's journal is empty.
	// Thus, here we iterate over stateObjects, to enable copies of copies
	for addr := range s.stateObjectsPending {
		if _, exist := state.stateObjects[addr]; !exist {
			state.stateObjects[addr] = s.stateObjects[addr].deepCopy(state)
		}
		state.stateObjectsPending[addr] = struct{}{}
	}
	for addr := range s.stateObjectsDirty {
		if _, exist := state.stateObjects[addr]; !exist {
			state.stateObjects[addr] = s.stateObjects[addr].deepCopy(state)
		}
		state.stateObjectsDirty[addr] = struct{}{}
	}

	return state
}

func (s *stateObject) deepCopy(db *StateDB) *stateObject {
	stateObject := newObject(db, s.address, s.data)
	if s.trie != nil {
		stateObject.trie = db.db.CopyTrie(s.trie)
	}

	stateObject.dirtyStorage = s.dirtyStorage.Copy()
	stateObject.originStorage = s.originStorage.Copy()
	stateObject.pendingStorage = s.pendingStorage.Copy()
	stateObject.suicided = s.suicided
	stateObject.deleted = s.deleted
	return stateObject
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

// SubBalance subtracts amount from the account associated with addr.
func (s *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObj(addr)
	if stateObject != nil {
		stateObject.SubBalance(amount)
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
			log.Fatal("Failed to commit dirty codes", "error", err)
		}
	}

	// The onleaf func is called _serially_, so we can reuse the same account
	// for unmarshalling every time.
	var account Account
	root, err := s.trie.Commit(func(_ [][]byte, _ []byte, leaf []byte, parent common.Hash) error {
		if err := encoding.Unmarshal(leaf, &account); err != nil {
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
	if err := encoding.Unmarshal(enc, data); err != nil {
		log.Error("Failed to decode state object", "addr", addr, "err", err)
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

	data, err := encoding.Marshal(obj)
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

func (s *StateDB) Snapshot() int {
	id := s.nextRevisionId
	s.nextRevisionId++
	s.validRevisions = append(s.validRevisions, revision{id, s.journal.length()})
	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (s *StateDB) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	idx := sort.Search(len(s.validRevisions), func(i int) bool {
		return s.validRevisions[i].id >= revid
	})
	if idx == len(s.validRevisions) || s.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := s.validRevisions[idx].journalIndex

	// Replay the journal to undo changes and remove invalidated snapshots
	s.journal.revert(s, snapshot)
	s.validRevisions = s.validRevisions[:idx]
}

func (s *StateDB) Prepare(thash, bhash common.Hash, ti int) {
	s.thash = thash
	s.bhash = bhash
	s.txIndex = ti
}

// Retrieve the balance from the given address or 0 if object not found
func (s *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return big.NewInt(0)
}

func (s *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}
	return 0
}

// Exist reports whether the given account exists in state.
func (s *StateDB) Exist(addr common.Address) bool {
	return s.getStateObject(addr) != nil
}

func (s *StateDB) CreateAccount(addr common.Address) {
	newObj, prev := s.createStateObject(addr)
	if prev != nil {
		newObj.setBalance(prev.data.Balance)
	}
}
