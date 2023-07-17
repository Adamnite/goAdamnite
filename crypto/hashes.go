package crypto

import (
	"crypto/sha512"
	"golang.org/x/crypto/ripemd160"
	"hash"
	"errors"
	"fmt"
	"crypto/sha256"
	"github.com/adamnite/goadamnite/utils"
	"github.com/adamnite/goadamnite/utils/math"
)

//A collection of various hashes for use
//Mainly can be used to generated hashes of messages for transactions, addresses, etc

const (
	Sha512_truncatedSize = sha512.Size256
	Sha512Size 			 = sha512.Size
	Ripemd160Size		 = ripemd160.Size

)

// Create a regular plain SHA-512 hash of a given plaintext message
func computeSHA512Hash(msg []byte) []byte {
	hash := sha512.Sum512(msg)
	return hash[:]
}

// Compute a truncated SHA-512 hash (truncating at the first half) for use when generating addresses.
func computeSHA512TruncatedHash(msg []byte) []byte {
	hash := sha512.Sum512(msg)
	return hash[:len(hash)/2]
}

// Generate a plain RIPEMD160 Hash 
func computeRIPEMD160Hash(msg []byte) []byte {
	hash := ripemd160.New()
	hash.Write(msg)
	return hash.Sum(nil)
}

// Verify stuff

func verifySHA512Hash(msg []byte, hash []byte) bool {
	computedHash := computeSHA512Hash(msg)
	return hmac.Equal(hash, computedHash)
}

func verifyRIPEMD160Hash(msg []byte, hash []byte) bool {
	computedHash := computeRIPEMD160Hash(msg)
	return hmac.Equal(hash, computedHash)