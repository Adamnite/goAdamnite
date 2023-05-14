package accounts

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/rand"
	"log"
	"math/big"

	"golang.org/x/crypto/ripemd160"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto/secp256k1"
)

type Account struct {
	Address    common.Address
	PublicKey  []byte
	PrivateKey []byte
	Balance    float64
}

func GenerateAccount() (*Account, error) {
	publicKey, privateKey, err := generateKeys()
	if err != nil {
		log.Printf("Account generation error: %s", err)
		return nil, err
	}

	return &Account{
		Address   : createAddress(publicKey),
		PublicKey : publicKey,
		PrivateKey: privateKey,
		Balance   : 0,
	}, nil
}

func (a *Account) Sign(data []byte) ([]byte, error) {
	signature, err := secp256k1.Sign(sha256Hash(data), a.PrivateKey)
	if err != nil {
		log.Printf("Signing error: %s", err)
		return nil, err
	}
	return signature, nil
}

func (a *Account) Verify(data []byte, signature []byte) bool {
	return secp256k1.VerifySignature(a.PublicKey, sha256Hash(data), signature)
}

func generateKeys() (rawPublicKey, rawPrivateKey []byte, err error) {
	privateKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		log.Printf("Keys generation error: %s", err)
		return rawPublicKey, rawPrivateKey, err
	}

	publicKey := privateKey.PublicKey

	rawPrivateKey = privateKey.D.Bytes()
	rawPublicKey = elliptic.Marshal(publicKey, publicKey.X, publicKey.Y)
	return
}

func createAddress(publicKey []byte) common.Address {
	var addr common.Address
	addr.SetBytes(b58encode(ripemd160Hash(sha512Hash(publicKey[1:]))))
	return addr
}

func b58encode(data []byte) []byte {
	const BASE58_CHARS = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	// convert big endian bytes to big int
	x := new(big.Int).SetBytes(data)

	// initialize
	r := new(big.Int)
	m := big.NewInt(58)
	zero := big.NewInt(0)
	s := ""

	// convert big int to string/
	for x.Cmp(zero) > 0 {
		// x, r = (x / 58, x % 58)
		x.QuoRem(x, m, r)
		// prepend ASCII character
		s = string(BASE58_CHARS[r.Int64()]) + s
	}

	return []byte(s)
}

func ripemd160Hash(data []byte) []byte {
	hasher := ripemd160.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

func sha256Hash(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

func sha512Hash(data []byte) []byte {
	hasher := sha512.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}