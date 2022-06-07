package rawdb

import (
	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/leveldb"
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
