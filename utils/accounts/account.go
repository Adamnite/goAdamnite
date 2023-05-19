package accounts

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"log"
	"math/big"

	"golang.org/x/crypto/ripemd160"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/crypto/secp256k1"
)

type Account struct {
	Address    common.Address
	PublicKey  []byte
	privateKey []byte
	Balance    *big.Int
}

func AccountFromPubBytes(pubKey []byte) Account {
	return Account{
		Address:   crypto.PubkeyByteToAddress(pubKey),
		PublicKey: pubKey,
	}
}
func AccountFromStorage(storagePoint string) (Account, error) {
	priv, err := crypto.LoadECDSA(storagePoint)
	if err != nil {
		return Account{}, err
	}
	return AccountFromPrivEcdsa(priv), nil
}
func AccountFromPrivEcdsa(privKey *ecdsa.PrivateKey) Account {
	publicKey := privKey.PublicKey

	return Account{
		Address:    createAddress(publicKey.X.Bytes()),
		PublicKey:  elliptic.Marshal(publicKey, publicKey.X, publicKey.Y),
		privateKey: privKey.D.Bytes(),
		Balance:    big.NewInt(0),
	}

}
func AccountFromPrivBytes(privKey []byte) Account {
	ePriv, err := crypto.ToECDSA(privKey)

	publicKey := ePriv.PublicKey
	if err != nil {
		return Account{}
	}
	return Account{
		Address:    createAddress(publicKey.X.Bytes()),
		PublicKey:  elliptic.Marshal(publicKey, publicKey.X, publicKey.Y),
		privateKey: privKey,
		Balance:    big.NewInt(0),
	}
}

func GenerateAccount() (*Account, error) {
	publicKey, privateKey, err := generateKeys()
	if err != nil {
		log.Printf("Account generation error: %s", err)
		return nil, err
	}

	return &Account{
		Address:    createAddress(publicKey),
		PublicKey:  publicKey,
		privateKey: privateKey,
		Balance:    big.NewInt(0),
	}, nil
}

// sign an array of data (bytes), and return a 65 byte array
func (a *Account) Sign(data []byte) ([]byte, error) {
	signature, err := secp256k1.Sign(sha256Hash(data), a.privateKey)
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
	addr.SetBytes(ripemd160Hash(sha512Hash(publicKey[1:])))
	return addr
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
