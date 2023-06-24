package database

import (
	"log"
	"sync"

	"github.com/adamnite/go-adamnite/utils"
	encoding "github.com/vmihailenco/msgpack/v5"
)

type WitnessDatabase struct {
	dbImpl         *Database
	nextRevisionId int
    mu             sync.RWMutex
}

func NewWitnessDatabase() (*WitnessDatabase, error) {
	db, err := New(getRandomDatabasePath()) // prepend witness to db path
	if err != nil {
		return nil, err
	}

	return &WitnessDatabase{
		dbImpl:         db,
		nextRevisionId: 0,
		mu:             sync.RWMutex{},
	}, nil
}

// Snapshot creates new snapshot of state database.
func (w *WitnessDatabase) Snapshot() int {
	id := w.nextRevisionId
	w.nextRevisionId++
	return id
}

// Restore restores state database to the revision with the specified ID.
func (s *WitnessDatabase) Restore(revisionId int) bool {
	// TODO: Implement revisions restoration
	return true
}

// CreateWitness adds new witness to the database.
func (s *WitnessDatabase) CreateWitness(witness *utils.Candidate) (bool, error) {
	hashData, err := encoding.Marshal(witness.Hash())
	if err != nil {
		log.Printf("[WitnessDatabase] Creating witness serialization error: %s", err)
		return false, err
	}

	witnessData, err := encoding.Marshal(witness)
	if err != nil {
		log.Printf("[WitnessDatabase] Creating witness serialization error: %s", err)
		return false, err
	}

	if err := s.dbImpl.Insert(hashData, witnessData); err != nil {
		log.Printf("[WitnessDatabase] Insertion error: %s", err)
		return false, err
	}
	return true, nil
}
