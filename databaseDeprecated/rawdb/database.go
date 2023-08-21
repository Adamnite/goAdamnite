package rawdb

import (
	"github.com/adamnite/go-adamnite/databaseDeprecated"
	"github.com/adamnite/go-adamnite/databaseDeprecated/leveldb"
	"github.com/adamnite/go-adamnite/databaseDeprecated/memorydb"
)

type AdamniteDB struct {
	adamnitedb.Database
}

func NewAdamniteDB(db adamnitedb.Database) adamnitedb.Database {
	return &AdamniteDB{
		Database: db,
	}
}

func NewAdamniteLevelDB(fileName string, cache int, handles int, readonly bool) (adamnitedb.Database, error) {
	db, err := leveldb.New(fileName, cache, handles, readonly)
	if err != nil {
		return nil, err
	}
	return NewAdamniteDB(db), nil
}

func NewMemoryDB() adamnitedb.Database {
	return NewAdamniteDB(memorydb.New())
}
