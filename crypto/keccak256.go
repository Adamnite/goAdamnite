package crypto

import (
	"hash"

	"github.com/adamnite/go-adamnite/common"
	"golang.org/x/crypto/sha3"
)

type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

// NewKeccakState creates a new KeccakState
func NewKeccakState() KeccakState {
	return sha3.NewLegacyKeccak256().(KeccakState)
}

// HashData hashes the provided data using the KeccakState and returns a 32 byte hash
func HashData(kh KeccakState, data []byte) (h common.Hash) {
	kh.Reset()
	kh.Write(data)
	kh.Read(h[:])
	return h
}

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	b := make([]byte, 32)
	d := NewKeccakState()
	for _, b := range data {
		d.Write(b)
	}
	d.Read(b)
	return b
}

// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
// converting it to an internal Hash data structure.
func Keccak256Hash(data ...[]byte) (h common.Hash) {
	d := NewKeccakState()
	for _, b := range data {
		d.Write(b)
	}
	d.Read(h[:])
	return h
}
