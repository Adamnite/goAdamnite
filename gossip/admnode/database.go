package admnode

import (
	"bytes"
	"encoding/binary"
	"os"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

// NodeDB is the node database, storing previously seen live nodes
type NodeDB struct {
	levelDB   *leveldb.DB
	isRunning sync.Once
	quit      chan struct{}
}

func OpenDB(path string) (*NodeDB, error) {
	if path == "" {
		return newMemoryNodeDB()
	}
	return newDiskNodeDB(path)
}

func newMemoryNodeDB() (*NodeDB, error) {
	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		return nil, err
	}

	return &NodeDB{levelDB: db, quit: make(chan struct{})}, nil
}

func newDiskNodeDB(path string) (*NodeDB, error) {
	options := &opt.Options{OpenFilesCacheCapacity: 5}
	db, err := leveldb.OpenFile(path, options)
	if _, isCorrupted := err.(*errors.ErrCorrupted); isCorrupted {
		db, err = leveldb.RecoverFile(path, nil)
	}
	if err != nil {
		return nil, err
	}

	curVersion := make([]byte, binary.MaxVarintLen64)
	curVersion = curVersion[:binary.PutVarint(curVersion, int64(dbVersion))]

	// Check the database version if exist
	byVersion, err := db.Get([]byte(dbVersionKey), nil)
	if err == nil {
		if !bytes.Equal(byVersion, curVersion) {
			db.Close()
			if err = os.RemoveAll(path); err != nil {
				return nil, err
			}
			return newDiskNodeDB(path)
		}
	} else if err == leveldb.ErrNotFound {
		if err := db.Put([]byte(dbVersionKey), curVersion, nil); err != nil {
			db.Close()
			return nil, err
		}
	}

	return &NodeDB{levelDB: db, quit: make(chan struct{})}, nil
}
