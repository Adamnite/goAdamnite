package statedb

import (
	"fmt"
	"io"
	"math/big"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/vmihailenco/msgpack/v5"
)

type Storage map[common.Hash]common.Hash

func (s Storage) String() (str string) {
	for key, value := range s {
		str += fmt.Sprintf("%X : %X \n", key, value)
	}
	return str
}

func (s Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range s {
		cpy[key] = value
	}
	return cpy
}

// stateObject represents an Adamnite account which is being modified.
//
// The usage pattern is as follows:
// First you need to obtain a state object.
// Account values can be accessed and modified through the object.
// Finally, call CommitTrie to write the modified storage trie into a database.
type stateObject struct {
	address  common.Address
	addrHash common.Hash
	data     Account
	db       *StateDB

	dbErr error

	trie Trie

	dirtyStorage   Storage
	pendingStorage Storage
	originStorage  Storage

	deleted  bool
	suicided bool
}

type Account struct {
	Nonce   uint64
	Balance *big.Int
	Root    common.Hash // merkle root of the storage trie
}

func newObject(db *StateDB, addr common.Address, data Account) *stateObject {
	if data.Balance == nil {
		data.Balance = new(big.Int)
	}

	if data.Root == (common.Hash{}) {
		data.Root = emptyRoot
	}

	return &stateObject{
		db:             db,
		address:        addr,
		addrHash:       common.BytesToHash(crypto.Sha512(addr[:])),
		data:           data,
		dirtyStorage:   make(Storage),
		originStorage:  make(Storage),
		pendingStorage: make(Storage),
	}
}

func (s *stateObject) setNonce(nonce uint64) {
	s.data.Nonce = nonce
}

// Returns the address of the contract/account
func (s *stateObject) Address() common.Address {
	return s.address
}

func (s *stateObject) Balance() *big.Int {
	return s.data.Balance
}

func (s *stateObject) Nonce() uint64 {
	return s.data.Nonce
}

func (s *stateObject) empty() bool {
	return s.data.Nonce == 0 && s.data.Balance.Sign() == 0
}

func (s *stateObject) SetBalance(amount *big.Int) {
	s.setBalance(amount)
}

func (s *stateObject) setBalance(amount *big.Int) {
	s.data.Balance = amount
}

func (s *stateObject) AddBalance(amount *big.Int) {
	s.SetBalance(new(big.Int).Add(s.Balance(), amount))
}

func (s *stateObject) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	s.SetBalance(new(big.Int).Sub(s.Balance(), amount))
}

func (s *stateObject) SetNonce(nonce uint64) {
	s.setNonce(nonce)
}

func (s *stateObject) finalise(prefetch bool) {
	slotsToPrefetch := make([][]byte, 0, len(s.dirtyStorage))
	for key, value := range s.dirtyStorage {
		s.pendingStorage[key] = value
		if value != s.originStorage[key] {
			slotsToPrefetch = append(slotsToPrefetch, common.CopyBytes(key[:]))
		}
	}

	// ToDO: Implement the prefetch logic

	if len(s.dirtyStorage) > 0 {
		s.dirtyStorage = make(Storage)
	}
}

func (s *stateObject) updateRoot(db Database) {
	if s.updateTrie(db) == nil {
		return
	}
	s.data.Root = s.db.trie.Hash()
}

// CommitTrie the storage trie of the object to db.
// This updates the trie root.
func (s *stateObject) CommitTrie(db Database) error {
	// If nothing changed, don't bother with hashing anything
	if s.updateTrie(db) == nil {
		return nil
	}
	if s.dbErr != nil {
		return s.dbErr
	}

	root, err := s.trie.Commit(nil)
	if err == nil {
		s.data.Root = root
	}
	return err
}

func (s *stateObject) updateTrie(db Database) Trie {
	s.finalise(false)

	if len(s.pendingStorage) == 0 {
		return s.trie
	}

	// Insert all the pending updates into the trie
	tr := s.getTrie(db)

	usedStorage := make([][]byte, 0, len(s.pendingStorage))
	for key, value := range s.pendingStorage {
		// Skip noop changes, persist actual changes
		if value == s.originStorage[key] {
			continue
		}
		s.originStorage[key] = value

		var v []byte
		if (value == common.Hash{}) {
			s.setError(tr.TryDelete(key[:]))
		} else {
			// Encoding []byte cannot fail, ok to ignore the error.
			v, _ = msgpack.Marshal(common.TrimLeftZeroes(value[:]))
			s.setError(tr.TryUpdate(key[:], v))
		}
		// ToDO: implement snapshot

		usedStorage = append(usedStorage, common.CopyBytes(key[:])) // Copy needed for closure
	}

	// ToDO: implement prefetcher

	if len(s.pendingStorage) > 0 {
		s.pendingStorage = make(Storage)
	}
	return tr
}

func (s *stateObject) getTrie(db Database) Trie {
	if s.trie == nil {
		// ToDo: implement prefetcher

		if s.trie == nil {
			var err error
			s.trie, err = db.OpenStorageTrie(s.addrHash, s.data.Root)
			if err != nil {
				s.trie, _ = db.OpenStorageTrie(s.addrHash, common.Hash{})
				s.setError(fmt.Errorf("can't create storage trie: %v", err))
			}
		}
	}
	return s.trie
}

// setError remembers the first non-nil error it is called with.
func (s *stateObject) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}