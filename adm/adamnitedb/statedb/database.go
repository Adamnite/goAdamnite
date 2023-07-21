package statedb

import (
	"fmt"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/golang/groupcache/lru"
)

const (
	// Number of codehash->size associations to keep.
	codeSizeCacheSize = 100000

	// Cache size granted for caching clean code.
	codeCacheSize = 64 * 1024 * 1024
)

type Database interface {
	// OpenTrie opens the main account trie.
	OpenTrie(root utils.Hash) (Trie, error)

	// OpenStorageTrie opens the storage trie of an account.
	OpenStorageTrie(addrHash, root utils.Hash) (Trie, error)

	// TrieDB retrieves the low level trie database used for data storage.
	TrieDB() *trie.Database
	CopyTrie(Trie) Trie
}

type Trie interface {
	GetKey([]byte) []byte
	TryGet(key []byte) ([]byte, error)
	TryUpdate(key, value []byte) error
	TryDelete(key []byte) error
	Hash() utils.Hash
	Commit(onleaf trie.LeafCallback) (utils.Hash, error)
	NodeIterator(startKey []byte) trie.NodeIterator
	Prove(key []byte, fromLevel uint, proofDb adamnitedb.AdamniteDBWriter) error
}

type cachingDB struct {
	db            *trie.Database
	codeSizeCache *lru.Cache
	codeCache     *fastcache.Cache
}

func NewDatabase(db adamnitedb.Database) Database {
	return NewDatabaseWithConfig(db, nil)
}

func NewDatabaseWithConfig(db adamnitedb.Database, config *trie.Config) Database {
	csc := lru.New(codeSizeCacheSize)
	return &cachingDB{
		db:            trie.NewDatabaseWithConfig(db, config),
		codeSizeCache: csc,
		codeCache:     fastcache.New(codeCacheSize),
	}
}

// OpenTrie opens the main account trie at a specific root hash.
func (db *cachingDB) OpenTrie(root utils.Hash) (Trie, error) {
	tr, err := trie.NewSecure(root, db.db)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

// OpenStorageTrie opens the storage trie of an account.
func (db *cachingDB) OpenStorageTrie(addrHash, root utils.Hash) (Trie, error) {
	tr, err := trie.NewSecure(root, db.db)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

// CopyTrie returns an independent copy of the given trie.
func (db *cachingDB) CopyTrie(t Trie) Trie {
	switch t := t.(type) {
	case *trie.SecureTrie:
		return t.Copy()
	default:
		panic(fmt.Errorf("unknown trie type %T", t))
	}
}

// TrieDB retrieves any intermediate trie-node caching layer.
func (db *cachingDB) TrieDB() *trie.Database {
	return db.db
}
