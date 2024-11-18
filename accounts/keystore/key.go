package keystore

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/adamnite/go-adamnite/accounts"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/google/uuid"
)

const (
	version = 1
)

type Key struct {
	Id         uuid.UUID
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

type keyStore interface {
	GetKey(addr common.Address, filename string, auth string) (*Key, error)
	StoreKey(filename string, k *Key, auth string) error
	JoinPath(filename string) string
}

type plainKeyJSON struct {
	Version    int    `json:"version"`
	Id         string `json:"id"`
	Address    string `json:"address"`
	PrivateKey string `json:"privkey"`
	CreatedAt  string `json:"createdat"`
}

type encryptedKeyJSON struct {
	Version    int        `json:"version"`
	Id         string     `json:"id"`
	Address    string     `json:"address"`
	CryptedKey CryptoJSON `json:"cryptedkey"`
	CreatedAt  string     `json:"createdat"`
}

type CryptoJSON struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams cipherparamsJSON       `json:"cipherparams"`
	KDF          string                 `json"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type cipherparamsJSON struct {
	IV string `json:"iv"`
}

func (k *Key) MarshalJSON() ([]byte, error) {
	jsonStruct := plainKeyJSON{
		version,
		k.Id.String(),
		hex.EncodeToString(k.Address[:]),
		hex.EncodeToString(crypto.FromECDSA(k.PrivateKey)),
		time.Now().String(),
	}

	return json.Marshal(jsonStruct)
}

func (k *Key) UnmarshalJSON(jBytes []byte) (err error) {
	keyJson := new(plainKeyJSON)

	if err = json.Unmarshal(jBytes, &keyJson); err != nil {
		return err
	}

	u := new(uuid.UUID)
	if *u, err = uuid.Parse(keyJson.Id); err != nil {
		return err
	}

	k.Id = *u

	addr, err := hex.DecodeString(keyJson.Address)
	if err != nil {
		return err
	}

	privKey, err := crypto.HexToECDSA(keyJson.PrivateKey)
	if err != nil {
		return err
	}

	k.Address = common.BytesToAddress(addr)
	k.PrivateKey = privKey

	return nil
}

func newKey(rand io.Reader) (*Key, error) {
	privkey, err := ecdsa.GenerateKey(crypto.S256(), rand)
	if err != nil {
		return nil, err
	}
	return newKeyFromECDSA(privkey), nil
}

func newKeyFromECDSA(privKey *ecdsa.PrivateKey) *Key {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(fmt.Sprintf("Could not create random uuid: %v", err))
	}

	key := &Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privKey.PublicKey),
		PrivateKey: privKey,
	}
	return key
}

func storeNewKey(ks keyStore, rand io.Reader, auth string) (*Key, accounts.Account, error) {
	key, err := newKey(rand)
	if err != nil {
		return nil, accounts.Account{}, err
	}

	a := accounts.Account{
		Address: key.Address,
		URL:     accounts.URL{ProtocolScheme: KeyStoreScheme, Path: ks.JoinPath(keyFileName(key.Address))},
	}

	if err := ks.StoreKey(a.URL.Path, key, auth); err != nil {
		zeroKey(key.PrivateKey)
		return nil, a, err
	}

	return key, a, err
}

func keyFileName(keyAddr common.Address) string {

	return fmt.Sprintf("ADAMNITE-%s", hex.EncodeToString(keyAddr[:]))
}

func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}
