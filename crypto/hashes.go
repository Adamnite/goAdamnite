package crypto

import (
	"crypto/sha512"
	"golang.org/x/crypto/ripemd160"
	"hash"
	"errors"
	"fmt"
	"crypto/sha256"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/math"
)



const (
	Sha512_truncatedSize = sha512.Size256
	Sha512Size 			 = sha512.Size
	Ripemd160Size		 = ripemd160.Size

)

// Computes the SHA-512 hash of the given message.
func computeSHA512Hash(msg []byte) []byte {
	hash := sha512.Sum512(msg)
	return hash[:]
}

// Computes the truncated SHA-512 hash (first half) of the given message.
func computeSHA512TruncatedHash(msg []byte) []byte {
	hash := sha512.Sum512(msg)
	return hash[:len(hash)/2]
}

// Computes the RIPEMD-160 hash of the given message.
func computeRIPEMD160Hash(msg []byte) []byte {
	hash := ripemd160.New()
	hash.Write(msg)
	return hash.Sum(nil)
}

// Verifies that the given hash matches the SHA-512 hash of the given message.
func verifySHA512Hash(msg []byte, hash []byte) bool {
	computedHash := computeSHA512Hash(msg)
	return hmac.Equal(hash, computedHash)
}

// Verifies that the given hash matches the RIPEMD-160 hash of the given message.
func verifyRIPEMD160Hash(msg []byte, hash []byte) bool {
	computedHash := computeRIPEMD160Hash(msg)
	return hmac.Equal(hash, computedHash)
}


