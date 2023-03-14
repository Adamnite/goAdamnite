// Implementing a binary merkle trie for account storage

package adamnitedb

import (
	"math/big"

	"github.com/adamnite/go-adamnite/common"
)

type Trie interface {
	Get() // Retrieve a value assicated with the given key from the trie
	Set() //
	Delete() //
	Update() //
}


// We are eliminating the statedb and storage tries. 

// ========================================================

// An account that will be stored inside the BMT
type AdmAccount struct {
	Balance 	big.Int
	Nonce		big.Int
	CodeHash	*common.Hash // Nil for EOA - Hash of the code stored in the offchain db
	StorageHash *common.Hash // Nil for EOA - The Merkle Root of stored data 
	Deleted 	bool 
}

func (a *AdmAccount) Get() {

}

func (a *AdmAccount) isRegularAccount() bool {
	return a.CodeHash == nil
}

func (a *AdmAccount) GetCode() {
	// Make a request to the offchain database with the .CodeHash 
	// And retrieve the code
}

func (a *AdmAccount) GetStorage() {
	// Make a request to the offchain database with the .StorageHash
}

type TxLog struct {

} 

type TxReceipt struct {
	Log 		TxLog
	TxHash 		common.Hash
	BlockHash	common.Hash
	BlockNumber	big.Int
}

// A transaction that will be stored inside the MBT
type AdmTransaction struct {
	Nonce		big.Int
	Ate			big.Int
	AteMax		big.Int
	Sender		common.Address
	Receiver	common.Address
	Amount 		big.Int // in micalli (the smallest denomination of nite)
	Message		[]byte
	Signature	[]byte
	Receipt 	TxReceipt
}

// The tracker for the witnesses that have served during the past rounds
type AdmWitness struct {
	RoundNumber		big.Int
	Witnessers 		[]byte
	BlocksHashes 	[]byte
}

type Storage map[common.Address]AdmAccount
