package keystore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/adamnite/go-adamnite/accounts"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/google/uuid"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

const (
	StandardScryptN = 1 << 20
	StandardScryptP = 1

	scryptR = 8
)

type keyStorePassphrase struct {
	scryptN     int
	scryptP     int
	keysDirPath string
}

func StoreKey(dir string, pwd string, scriptN int, scriptP int) (accounts.Account, error) {
	_, a, err := storeNewKey(&keyStorePassphrase{scryptN: scriptN, scryptP: scriptP, keysDirPath: dir}, rand.Reader, pwd)
	return a, err
}

func GetKey(addr common.Address, filename string, pwd string) (*Key, error) {
	keyjson, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	key, err := DecryptKey(keyjson, pwd)
	if err != nil {
		return nil, err
	}

	if key.Address != addr {
		return nil, fmt.Errorf("key content mismatch: have account %x, want %x", key.Address, addr)
	}

	return key, nil
}

func (ks keyStorePassphrase) StoreKey(filename string, key *Key, pwd string) error {
	keyjson, err := EncryptKey(key, pwd, ks.scryptN, ks.scryptP)
	if err != nil {
		return err
	}

	tmpName, err := writeTemporaryKeyFile(filename, keyjson)
	if err != nil {
		return err
	}

	_, err = ks.GetKey(key.Address, tmpName, pwd)
	if err != nil {
		msg := "An error was encountered when saving and verifying the keystore file. \n" +
			"This indicates that the keystore is corrupted. \n" +
			"The corrupted file is stored at \n%v\n" +
			"Please file a ticket at:\n\n" +
			"The error was : %s"
		//lint:ignore ST1005 This is a message for the user
		return fmt.Errorf(msg, tmpName, err)
	}
	return os.Rename(tmpName, filename)
}

func (ks keyStorePassphrase) GetKey(addr common.Address, filename string, pwd string) (*Key, error) {
	keyjson, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	key, err := DecryptKey(keyjson, pwd)
	if err != nil {
		return nil, err
	}

	if key.Address != addr {
		return nil, fmt.Errorf("key content mismatch: have account %x, want %x", key.Address, addr)
	}

	return key, nil
}

func (ks keyStorePassphrase) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(ks.keysDirPath, filename)
}

func DecryptKey(keyjson []byte, pwd string) (*Key, error) {
	m := make(map[string]interface{})
	if err := json.Unmarshal(keyjson, &m); err != nil {
		return nil, err
	}

	var (
		keyBytes, keyId []byte
		err             error
	)
	if version := toInt(m["version"]); version == 1 {
		k := new(encryptedKeyJSON)
		if err := json.Unmarshal(keyjson, k); err != nil {
			return nil, err
		}
		keyBytes, keyId, err = decryptKey(k, pwd)
	}

	if err != nil {
		return nil, err
	}

	key := crypto.ToECDSAUnsafe(keyBytes)
	id, err := uuid.FromBytes(keyId)
	if err != nil {
		return nil, err
	}
	return &Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(key.PublicKey),
		PrivateKey: key,
	}, nil
}

func EncryptKey(key *Key, pwd string, scriptN, scriptP int) ([]byte, error) {
	keyBytes := math.PaddedBigBytes(key.PrivateKey.D, 32)
	cryptedKey, err := EncryptData(keyBytes, []byte(pwd), scriptN, scriptP)
	if err != nil {
		return nil, err
	}

	encryptedKeyJSON := encryptedKeyJSON{
		1,
		key.Id.String(),
		hex.EncodeToString(key.Address[:]),
		cryptedKey,
		time.Now().String(),
	}

	return json.Marshal(encryptedKeyJSON)
}

func EncryptData(data, pwd []byte, scryptN, scriptP int) (CryptoJSON, error) {
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic("reading random salt failed: " + err.Error())
	}

	derivedKey, err := scrypt.Key(pwd, salt, scryptN, scryptR, scriptP, 32)
	if err != nil {
		return CryptoJSON{}, err
	}

	encryptKey := derivedKey[:16]

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("reading random iv failed: " + err.Error())
	}

	cipherText, err := aesCTRXOR(encryptKey, data, iv)
	if err != nil {
		return CryptoJSON{}, err
	}

	mac := crypto.Keccak256(derivedKey[16:32], cipherText)
	scryptParamsJson := make(map[string]interface{}, 5)
	scryptParamsJson["n"] = scryptN
	scryptParamsJson["p"] = scriptP
	scryptParamsJson["r"] = scryptR
	scryptParamsJson["keylen"] = 32
	scryptParamsJson["salt"] = hex.EncodeToString(salt)

	cipherParamsJSON := cipherparamsJSON{
		IV: hex.EncodeToString(iv),
	}

	cryptedKey := CryptoJSON{
		Cipher:       "adam-aes-128",
		CipherText:   hex.EncodeToString(cipherText),
		CipherParams: cipherParamsJSON,
		KDF:          "scrypt",
		KDFParams:    scryptParamsJson,
		MAC:          hex.EncodeToString(mac),
	}

	return cryptedKey, nil
}

func DecryptData(cryptedKey CryptoJSON, pwd string) ([]byte, error) {
	if cryptedKey.Cipher != "adam-aes-128" {
		return nil, fmt.Errorf("cipher not supported: %v", cryptedKey.Cipher)
	}

	mac, err := hex.DecodeString(cryptedKey.MAC)
	if err != nil {
		return nil, err
	}

	iv, err := hex.DecodeString(cryptedKey.CipherParams.IV)
	if err != nil {
		return nil, err
	}

	cipherText, err := hex.DecodeString(cryptedKey.CipherText)
	if err != nil {
		return nil, err
	}

	derivedKey, err := getKDFKey(cryptedKey, pwd)
	if err != nil {
		return nil, err
	}

	calculatedMAC := crypto.Keccak256(derivedKey[16:32], cipherText)
	if !bytes.Equal(calculatedMAC, mac) {
		return nil, errors.New("could not decrypt key with given password")
	}

	plainText, err := aesCTRXOR(derivedKey[:16], cipherText, iv)
	if err != nil {
		return nil, err
	}
	return plainText, err
}

func decryptKey(encryptedKey *encryptedKeyJSON, pwd string) (keyBytes []byte, keyId []byte, err error) {
	if encryptedKey.Version != 1 {
		return nil, nil, fmt.Errorf("version not supported: %v", encryptedKey.Version)
	}

	keyUUID, err := uuid.Parse(encryptedKey.Id)
	if err != nil {
		return nil, nil, err
	}
	keyId = keyUUID[:]
	plainText, err := DecryptData(encryptedKey.CryptedKey, pwd)
	if err != nil {
		return nil, nil, err
	}
	return plainText, keyId, err
}

func getKDFKey(cryptedKeyJson CryptoJSON, pwd string) ([]byte, error) {
	pwdArray := []byte(pwd)
	salt, err := hex.DecodeString(cryptedKeyJson.KDFParams["salt"].(string))
	if err != nil {
		return nil, err
	}

	keyLen := toInt(cryptedKeyJson.KDFParams["keylen"])

	if cryptedKeyJson.KDF == "scrypt" {
		n := toInt(cryptedKeyJson.KDFParams["n"])
		r := toInt(cryptedKeyJson.KDFParams["r"])
		p := toInt(cryptedKeyJson.KDFParams["p"])
		return scrypt.Key(pwdArray, salt, n, r, p, keyLen)
	} else if cryptedKeyJson.KDF == "pbkdf2" {
		c := toInt(cryptedKeyJson.KDFParams["c"])
		prf := cryptedKeyJson.KDFParams["prf"].(string)
		if prf != "hmac-sha256" {
			return nil, fmt.Errorf("unsupported PBKDF2 PRF: %s", prf)
		}
		key := pbkdf2.Key(pwdArray, salt, c, keyLen, sha256.New)
		return key, nil
	}

	return nil, fmt.Errorf("unsupported KDF: %s", cryptedKeyJson.KDF)
}

func toInt(x interface{}) int {
	res, ok := x.(int)
	if !ok {
		res = int(x.(float64))
	}
	return res
}

func aesCBCDecrypt(key, cipherText, iv []byte) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypter := cipher.NewCBCDecrypter(aesBlock, iv)
	paddedPlaintext := make([]byte, len(cipherText))
	decrypter.CryptBlocks(paddedPlaintext, cipherText)
	plaintext := pkcs7Unpad(paddedPlaintext)
	if plaintext == nil {
		return nil, errors.New("could not decrypt key with given password")
	}
	return plaintext, err
}

func pkcs7Unpad(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}

	padding := in[len(in)-1]
	if int(padding) > len(in) || padding > aes.BlockSize {
		return nil
	} else if padding == 0 {
		return nil
	}

	for i := len(in) - 1; i > len(in)-int(padding)-1; i-- {
		if in[i] != padding {
			return nil
		}
	}
	return in[:len(in)-int(padding)]
}

func aesCTRXOR(key, data, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(data))
	stream.XORKeyStream(outText, data)
	return outText, err
}

func writeTemporaryKeyFile(file string, content []byte) (string, error) {
	// Create the keystore directory with appropriate permissions
	// in case it is not present yet.
	const dirPerm = 0700
	if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
		return "", err
	}
	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	f.Close()
	return f.Name(), nil
}
