package database

import (
	"errors"
	"log"
	"math/big"
	"sync"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils/accounts"
	encoding "github.com/vmihailenco/msgpack/v5"
)

var (
	ErrorNonExistingAddress = errors.New("Address does not exist")
)

type StateDatabase struct {
	dbImpl         *Database
	nextRevisionId int
    mu             sync.RWMutex
}

func NewStateDatabase() (*StateDatabase, error) {
	db, err := New(getRandomDatabasePath()) // prepend state to db path
	if err != nil {
		return nil, err
	}

	return &StateDatabase{
		dbImpl:         db,
		nextRevisionId: 0,
		mu:             sync.RWMutex{},
	}, nil
}

// Snapshot creates new snapshot of state database.
func (s *StateDatabase) Snapshot() int {
	id := s.nextRevisionId
	s.nextRevisionId++
	return id
}

// Restore restores state database to the revision with the specified ID.
func (s *StateDatabase) Restore(revisionId int) bool {
	// TODO: Implement revisions restoration
	return true
}

// CreateAccount adds new account to the database.
func (s *StateDatabase) CreateAccount(address common.Address, account *accounts.Account) (bool, error) {
	addressData, err := encoding.Marshal(address)
	if err != nil {
		log.Printf("[StateDatabase] Creating account serialization error: %s", err)
		return false, err
	}

	accountData, err := encoding.Marshal(account)
	if err != nil {
		log.Printf("[StateDatabase] Creating account serialization error: %s", err)
		return false, err
	}

	if err := s.dbImpl.Insert(addressData, accountData); err != nil {
		log.Printf("[StateDatabase] Insertion error: %s", err)
		return false, err
	}
	return true, nil
}

// AccountExists checks whether the account with specified address exists.
func (s *StateDatabase) AccountExists(address common.Address) (bool, error) {
	addressData, err := encoding.Marshal(address)
	if err != nil {
		log.Printf("[StateDatabase] Account exists serialization error: %s", err)
		return false, err
	}

	value, err := s.dbImpl.Get(addressData)
	if err != nil {
		log.Printf("[StateDatabase] Account exists check error: %s", err)
		return false, err
	}
	return value == nil, nil
}

// GetBalance gets the balance of the account with specified address.
func (s *StateDatabase) GetBalance(address common.Address) (*big.Int, error) {
	addressData, err := encoding.Marshal(address)
	if err != nil {
		log.Printf("[StateDatabase] Get balance serialization error: %s", err)
		return nil, err
	}

	value, err := s.dbImpl.Get(addressData)
	if err != nil {
		log.Printf("[StateDatabase] Get balance error: %s", err)
		return nil, err
	}

	if value == nil {
		return nil, ErrorNonExistingAddress
	}

	var account accounts.Account
	if err := encoding.Unmarshal(value, &account); err != nil {
		log.Printf("[StateDatabase] Get balance error: %s", err)
		return nil, err
	}

	return account.Balance, nil
}

// Transfer transfers the specified amount from sender to receiver.
func (s *StateDatabase) Transfer(sender common.Address, receiver common.Address, amount *big.Int) {
	s.SubBalance(sender, amount)
	s.AddBalance(receiver, amount)
}

func (s *StateDatabase) AddBalance(address common.Address, amount *big.Int) (bool, error) {
	addressData, err := encoding.Marshal(address)
	if err != nil {
		log.Printf("[StateDatabase] Add balance serialization error: %s", err)
		return false, err
	}

	value, err := s.dbImpl.Get(addressData)
	if err != nil {
		log.Printf("[StateDatabase] Add balance error: %s", err)
		return false, err
	}

	if value == nil {
		return false, ErrorNonExistingAddress
	}

	var account accounts.Account
	if err := encoding.Unmarshal(value, &account); err != nil {
		log.Printf("[StateDatabase] Add balance error: %s", err)
		return false, err
	}

	account.Balance.Add(account.Balance, amount)

	updatedAccountData, err := encoding.Marshal(account)
	if err != nil {
		log.Printf("[StateDatabase] Add balance serialization error: %s", err)
		return false, err
	}

	if err := s.dbImpl.Insert(addressData, updatedAccountData); err != nil {
		log.Printf("[StateDatabase] Add balance insertion error: %s", err)
		return false, err
	}
	return true, nil
}

func (s *StateDatabase) SubBalance(address common.Address, amount *big.Int) (bool, error) {
	addressData, err := encoding.Marshal(address)
	if err != nil {
		log.Printf("[StateDatabase] Subtract balance serialization error: %s", err)
		return false, err
	}

	value, err := s.dbImpl.Get(addressData)
	if err != nil {
		log.Printf("[StateDatabase] Subtract balance error: %s", err)
		return false, err
	}

	if value == nil {
		return false, ErrorNonExistingAddress
	}

	var account accounts.Account
	if err := encoding.Unmarshal(value, &account); err != nil {
		log.Printf("[StateDatabase] Subtract balance error: %s", err)
		return false, err
	}

	account.Balance.Sub(account.Balance, amount)

	updatedAccountData, err := encoding.Marshal(account)
	if err != nil {
		log.Printf("[StateDatabase] Add balance serialization error: %s", err)
		return false, err
	}

	if err := s.dbImpl.Insert(addressData, updatedAccountData); err != nil {
		log.Printf("[StateDatabase] Subtract balance insertion error: %s", err)
		return false, err
	}
	return true, nil
}

func (s *StateDatabase) GetNonce(address common.Address) uint64 {
	// TODO: Account should contain nonce
	return 0
}

func (s *StateDatabase) SetNonce(address common.Address, nonce uint64) (bool, error) {
	// TODO: Account should contain nonce
	return false, nil
}