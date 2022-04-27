package enode

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

type DB struct {
	levelDB *leveldb.DB
	quit    chan struct{}
}

// OpenDB opens a node database for storing and retrieving infors about known peers in the network.
// If no path is given an in-memory, temporary database is constructed.
func OpenDB(path string) (*DB, error) {
	if path == "" {
		return newMemoryDB()
	}

	return newPersistentDB(path)
}

func newMemoryDB() (*DB, error) {
	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		return nil, err
	}

	return &DB{levelDB: db, quit: make(chan struct{})}, nil
}

func newPersistentDB(path string) (*DB, error) {
	return nil, fmt.Errorf("node db was not implemented yet")
}
