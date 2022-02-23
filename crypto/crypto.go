//Copyright Archie Chaudhury and Tsiamfei 2022
//Package Crypto is heavily inspired by Go-Ethereum's implmentation,
//which was authored by Jeffery Wilcke and others.

package crypto

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	csha512 "crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"math/big"
	"os"

	"golang.org/x/crypto/ripemd160"

	"github.com/adamnite/go-adamnite/common"
)

var (
	secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

// GenerateKey generates a new private key.
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(Secp256k1(), rand.Reader)
}

type mainState interface {
	hash.Hash
	Read([]byte) (int, error)
}

func newState() mainState {
	return csha512.New512_256().(mainState)
}

func sha512(data ...[]byte) []byte {
	b := make([]byte, 64)
	d := newState()
	for _, b := range data {
		d.Write(b)
	}
	d.Read(b)
	return b
}

func NewRipemd160State() hash.Hash {
	return ripemd160.New()
}

func Ripemd160Hash(data ...[]byte) []byte {
	hasher := NewRipemd160State()

	for _, b := range data {
		hasher.Write(b)
	}

	hashBytes := hasher.Sum(nil)
	return hashBytes
}

//String Function can be used in cases where a message needs to be hashed
func NewRipemd160String(s string) []byte {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return nil
	}

	return Ripemd160Hash(bytes)
}

// func CreateAddress(a common.Address, nonce uint64) common.Address {
//For now assuming we use recursive-length-prefix for the POC implementation
//Probably will want to use a seperate implementation or create our own
//considering Ethereum's problems with securely encoding values in smart contracts.
// data, _ := rlp.EncodeToBytes([]interface{}{a, nonce})
// return common.BytesToAddress(Ripemd160Hash(sha512(data))[12:])
// }
//Creates Address, given an initial salt for encoding
func CreateAddress(a common.Address, nonce uint64) common.Address {
  	return common.BytesToAddress(Ripemd160Hash(Keccak256([]byte{0xff}, b.Bytes(), salt[:], inithash)[12:])
  }
// ToECDSA creates a private key with the given D value.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	return toECDSA(d, true)
}

// ToECDSAUnsafe blindly converts a binary blob to a private key. It should almost
// never be used unless you are sure the input is valid and want to avoid hitting
// errors due to bad origin encoding (0 prefixes cut off).
func ToECDSAUnsafe(d []byte) *ecdsa.PrivateKey {
	priv, _ := toECDSA(d, false)
	return priv
}

// HexToECDSA parses a secp256k1 private key.
func HexToECDSA(hexKey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexKey)
	if byteErr, ok := err.(hex.InvalidByteError); ok {
		return nil, fmt.Errorf("invalid hex character %q in private key", byte(byteErr))
	} else if err != nil {
		return nil, errors.New("invalid hex data for private key")
	}

	return ToECDSA(b)
}

// LoadECDSA loads a secp256k1 private key from the given file.
func LoadECDSA(file string) (*ecdsa.PrivateKey, error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	r := bufio.NewReader(fd)
	buf := make([]byte, 64)
	n, err := readASCII(buf, r)
	if err != nil {
		return nil, err
	} else if n != len(buf) {
		return nil, fmt.Errorf("key file too short, want 64 hex characters")
	}
	if err := checkKeyFileEnd(r); err != nil {
		return nil, err
	}

	return HexToECDSA(string(buf))
}

// toECDSA creates a private key with the given D value. The strict parameter
// controls whether the key's length should be enforced at the curve size or
// it can also accept legacy encodings (0 prefixes).
func toECDSA(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = Secp256k1()

	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}

	priv.D = new(big.Int).SetBytes(d)

	if priv.D.Cmp(secp256k1N) >= 0 {
		return nil, errors.New("invalid private key, >=N")
	}

	if priv.D.Sign() <= 0 {
		return nil, errors.New("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}

	return priv, nil
}

// readASCII reads into 'buf', stopping when the buffer is full or
// when a non-printable control character is encountered.
func readASCII(buf []byte, r *bufio.Reader) (n int, err error) {
	for ; n < len(buf); n++ {
		buf[n], err = r.ReadByte()
		switch {
		case err == io.EOF || buf[n] < '!':
			return n, nil
		case err != nil:
			return n, err
		}
	}
	return n, nil
}

// checkKeyFileEnd skips over additional newlines at the end of a key file.
func checkKeyFileEnd(r *bufio.Reader) error {
	for i := 0; ; i++ {
		b, err := r.ReadByte()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case b != '\n' && b != '\r':
			return fmt.Errorf("invalid character %q at end of key file", b)
		case i >= 2:
			return errors.New("key file too long, want 64 hex characters")
		}
	}
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}

	return elliptic.Marshal(Secp256k1(), pub.X, pub.Y)
}

func PubkeyToAddress(p ecdsa.PublicKey) common.Address {
	pubBytes := FromECDSAPub(&p)
	return common.BytesToAddress(Ripemd160Hash(sha512(pubBytes)))
}
