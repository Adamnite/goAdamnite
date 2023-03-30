package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/crypto/secp256k1"
)

func Secp256k1() elliptic.Curve {
	return secp256k1.S256()
}

// Recover Public Key from signature. Hash mush be 32 bytes, signature must be 65 bytes.
func Recover(hash, signature []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, signature)
}

func Sign(dataHash []byte, prv *ecdsa.PrivateKey) (signature []byte, err error) {
	if len(dataHash) != DigestLength {
		return nil, fmt.Errorf("hash length should be %d bytes (%d)", DigestLength, len(dataHash))
	}
	secure_key := math.PaddedBigBytes(prv.D, prv.Params().BitSize/16)
	defer zeroBytes(secure_key)
	return secp256k1.Sign(dataHash, secure_key)
}

func VerifySignature(public_key, dataHash, signature []byte) bool {
	return secp256k1.VerifySignature(public_key, dataHash, signature)
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func CompressPubkey(pubkey *ecdsa.PublicKey) []byte {
	return secp256k1.CompressPubkey(pubkey.X, pubkey.Y)
}

// DecompressPubkey parses a public key in the 33-byte compressed format.
func DecompressPubkey(pubkey []byte) (*ecdsa.PublicKey, error) {
	x, y := secp256k1.DecompressPubkey(pubkey)
	if x == nil {
		return nil, fmt.Errorf("invalid public key")
	}
	return &ecdsa.PublicKey{X: x, Y: y, Curve: S256()}, nil
}

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return secp256k1.S256()
}
