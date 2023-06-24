package database

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/adamnite/go-adamnite/adm/merkle"

	"github.com/syndtr/goleveldb/leveldb"
)

type Database struct {
	Path string             // path to LevelDB instance
	impl *leveldb.DB        // LevelDB instance
	tree *merkle.MerkleTree // Merkle Tree instance
}

func New(path string) (*Database, error) {
	impl, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Printf("[Database] Opening file error: %s", err)
		return nil, err
	}

	db := &Database{
		Path: path,
		impl: impl,
		tree: merkle.NewEmptyTree(),
	}
	return db, nil
}

func (db *Database) Close() error {
	return db.impl.Close()
}

// Get gets value by specific key from the key-value store.
func (db *Database) Get(key []byte) ([]byte, error) {
	value, err := db.impl.Get(key, nil)
	if err != nil {
		log.Printf("[Database] Get value error: %s", err)
		return nil, err
	}
	return value, nil
}

// Insert inserts the given value into the key-value store.
func (db *Database) Insert(key []byte, value []byte) error {
	return db.impl.Put(key, value, nil)
}

// Delete removes the given value from the key-value store.
func (db *Database) Delete(key []byte) error {
	return db.impl.Delete(key, nil)
}

func getRandomDatabasePath() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("db-%d", rand.Int())
}