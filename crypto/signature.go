package crypto

import (
	"crypto/elliptic"

	"github.com/adamnite/go-adamnite/crypto/secp256k1"
)

func Secp256k1() elliptic.Curve {
	return secp256k1.S256()
}
