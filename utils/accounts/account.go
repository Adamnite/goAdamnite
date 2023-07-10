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
	"github.com/adamnite/go-adamnite/crypto/ecies"
	"github.com/adamnite/go-adamnite/crypto/secp256k1"
)

type Account struct {
	Address    common.Address //TODO: once GetAddress is the more commonly used item, we should make this private(lowercase)
	PublicKey  []byte
	privateKey []byte
	Balance    *big.Int
}

func AccountFromPubBytes(pubKey []byte) Account {
	//TODO: should add an error for if the pubKey is invalid
	return Account{
		Address:   crypto.PubkeyByteToAddress(pubKey),
		PublicKey: pubKey,
	}
}
func AccountFromStorage(storagePoint string) (*Account, error) {
	priv, err := crypto.LoadECDSA(storagePoint)
	if err != nil {
		return nil, err
	}
	return AccountFromPrivEcdsa(priv), nil
}

func AccountFromPrivEcdsa(privKey *ecdsa.PrivateKey) *Account {
	publicKey := privKey.PublicKey
	return &Account{
		Address:    createAddress(publicKey.X.Bytes()),
		PublicKey:  elliptic.Marshal(publicKey, publicKey.X, publicKey.Y),
		privateKey: privKey.D.Bytes(),
		Balance:    big.NewInt(0),
	}

}
func AccountFromPrivBytes(privKey []byte) (Account, error) {
	ePriv, err := crypto.ToECDSA(privKey)
	if err != nil {
		return Account{}, err
	}
	publicKey := ePriv.PublicKey
	return Account{
		Address:    createAddress(publicKey.X.Bytes()),
		PublicKey:  elliptic.Marshal(publicKey, publicKey.X, publicKey.Y),
		privateKey: privKey,
		Balance:    big.NewInt(0),
	}, nil
}

func GenerateAccount() (*Account, error) {
	privateKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		log.Printf("Account generation error: %s", err)
		return nil, err
	}
	if len(privateKey.D.Bytes()) != 32 {
		//sometimes ecdsa generates private keys or length 32. This is the easiest way to solve that.
		return GenerateAccount()
	}
	return AccountFromPrivEcdsa(privateKey), nil
}

// USE THIS OVER CALLING THE ADDRESS DIRECTLY
func (a *Account) GetAddress() common.Address {
	//this way we only store the address as its needed. Since it can be calculated when needed, no need to send it over networks.
	if a.Address == common.BytesToAddress([]byte{}) {
		a.Address = createAddress(a.PublicKey)
	}
	return a.Address
}

// ONLY USE THIS IN CLI! THIS IS NOT NORMALLY MEANT TO BE USED!
func (a *Account) GetPrivateB58() string {
	return crypto.B58encode(a.privateKey)
}

func (a *Account) Store(storagePoint string) error {
	k, err := crypto.ToECDSA(a.privateKey)
	if err != nil {
		return err
	}
	return crypto.SaveECDSA(storagePoint, k)
}

// sign data, and return a 65 byte array. Data can be most interface types
func (a *Account) Sign(data interface{}) ([]byte, error) {
	signature, err := secp256k1.Sign(toHashedBytes(data), a.privateKey)
	if err != nil {
		log.Printf("Signing error: %s", err)
		return nil, err
	}
	return signature, nil
}

// verify a signature is signed by this account, for the data passed
func (a *Account) Verify(data interface{}, signature []byte) bool {
	return secp256k1.VerifySignature(a.PublicKey, toHashedBytes(data), signature[:64])
}

// encrypt a message
func (a *Account) Encrypt(msg []byte) ([]byte, error) {
	pubKey := ecdsa.PublicKey{
		Curve: secp256k1.S256(),
	}
	pubKey.X, pubKey.Y = elliptic.Unmarshal(pubKey, a.PublicKey)

	eciesPub := ecies.ImportECDSAPublic(&pubKey)
	return ecies.Encrypt(rand.Reader, eciesPub, msg, nil, nil)
}

// decrypt the message
func (a *Account) Decrypt(msg []byte) ([]byte, error) {
	ecdsaKey, err := crypto.ToECDSA(a.privateKey)
	if err != nil {
		return nil, err
	}
	priv := ecies.ImportECDSA(ecdsaKey)
	return priv.Decrypt(msg, nil, nil)
}

// types that have a hash method that returns common.Hash
type commonHashAble interface{ Hash() common.Hash }

// types that have a hash method that returns bytes
type hashAble interface{ Hash() []byte }

type hasGetBytes interface{ Bytes() []byte }

// for handling multiple interface types and getting the hash
func toHashedBytes(data interface{}) []byte {
	var dataBytes []byte = []byte{}
	switch v := data.(type) {
	case commonHashAble:
		dataBytes = v.Hash().Bytes()
	case hashAble:
		dataBytes = v.Hash()
	case hasGetBytes:
		dataBytes = v.Bytes()
	case string:
		dataBytes = []byte(v)
	case []byte:
		dataBytes = v
	case []commonHashAble:
		for _, a := range v {
			dataBytes = append(dataBytes, a.Hash().Bytes()...)
		}
	case []hashAble:
		for _, a := range v {
			dataBytes = append(dataBytes, a.Hash()...)
		}
	}

	//insure that the dataBytes are of the correct length
	if len(dataBytes) != 32 {
		dataBytes = sha256Hash(dataBytes)
	}
	return dataBytes
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
