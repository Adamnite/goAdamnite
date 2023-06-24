package database

import (
	"log"
	"sync"

	"github.com/adamnite/go-adamnite/utils"
	encoding "github.com/vmihailenco/msgpack/v5"
)

type TransactionDatabase struct {
	dbImpl         *Database
	nextRevisionId int
    mu             sync.RWMutex
}

func NewTransactionDatabase() (*TransactionDatabase, error) {
	db, err := New(getRandomDatabasePath()) // prepend transaction to db path
	if err != nil {
		return nil, err
	}

	return &TransactionDatabase{
		dbImpl:         db,
		nextRevisionId: 0,
		mu:             sync.RWMutex{},
	}, nil
}

// Snapshot creates new snapshot of state database.
func (t *TransactionDatabase) Snapshot() int {
	id := t.nextRevisionId
	t.nextRevisionId++
	return id
}

// Restore restores state database to the revision with the specified ID.
func (s *TransactionDatabase) Restore(revisionId int) bool {
	// TODO: Implement revisions restoration
	return true
}

// CreateTransaction adds new transaction to the database.
func (s *TransactionDatabase) CreateTransaction(transaction *utils.Transaction) (bool, error) {
	hashData, err := encoding.Marshal(transaction.Hash())
	if err != nil {
		log.Printf("[TransactionDatabase] Creating transaction serialization error: %s", err)
		return false, err
	}

	transactionData, err := encoding.Marshal(transaction)
	if err != nil {
		log.Printf("[TransactionDatabase] Creating transaction serialization error: %s", err)
		return false, err
	}

	if err := s.dbImpl.Insert(hashData, transactionData); err != nil {
		log.Printf("[TransactionDatabase] Insertion error: %s", err)
		return false, err
	}
	return true, nil
}
