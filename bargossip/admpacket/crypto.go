package admpacket

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"hash"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/utils/math"
	"github.com/adamnite/go-adamnite/crypto"
	"golang.org/x/crypto/hkdf"
)

const (
	aesKeySize = 16
)

// encryptGCM encrypts plaintext(pt) using AES-GCM with the given key and nonce.
// The ciphertext is appended to dest, which must not overlap with plaintext.
func encryptGCM(dest, key, nonce, pt, authdata []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(fmt.Errorf("cannot create block cipher: %v", err))
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		panic(fmt.Errorf("cannot create GCM: %v", err))
	}

	return aesgcm.Seal(dest, nonce, pt, authdata), nil
}

// decryptGCM decryptes ct using AES-GCM with the given key and nonce.
func decryptGCM(key, nonce, ct, authdata []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("cannot create block cipher: %v", err)
	}
	if len(nonce) != nonceSize {
		return nil, fmt.Errorf("invalid GCM nonce size: %d", len(nonce))
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		panic(fmt.Errorf("cannot create GCM: %v", err))
	}
	pt := make([]byte, 0, len(ct))
	return aesgcm.Open(pt, nonce, ct, authdata)
}

func EncodePubkey(key *ecdsa.PublicKey) []byte {
	switch key.Curve {
	case crypto.S256():
		return crypto.CompressPubkey(key)
	default:
		panic("unsupported curve " + key.Curve.Params().Name)
	}
}

func DecodePubkey(curve elliptic.Curve, encKey []byte) (*ecdsa.PublicKey, error) {
	switch curve {
	case crypto.S256():
		if len(encKey) != 33 {
			return nil, errors.New("wrong size public key data")
		}
		return crypto.DecompressPubkey(encKey)
	default:
		return nil, fmt.Errorf("unsupported curve %s", curve.Params().Name)
	}
}

func getInputHash(h hash.Hash, destID admnode.NodeID, data, pubkey []byte) []byte {
	h.Reset()
	h.Write(destID[:])
	h.Write(data)
	h.Write(pubkey)
	return h.Sum(nil)
}

func makeSignature(hash hash.Hash, key *ecdsa.PrivateKey, destID admnode.NodeID, data, pubkey []byte) ([]byte, error) {
	input := getInputHash(hash, destID, data, pubkey)
	switch key.Curve {
	case crypto.S256():
		sig, err := crypto.Sign(input, key)
		if err != nil {
			return nil, err
		}
		return sig[:len(sig)-1], nil
	default:
		return nil, fmt.Errorf("unsupported curve %s", key.Curve.Params().Name)
	}
}

func verifySignature(hash hash.Hash, sig []byte, n *admnode.GossipNode, destID admnode.NodeID, data, pubkey []byte) error {
	input := getInputHash(hash, destID, data, pubkey)
	if !crypto.VerifySignature(crypto.FromECDSAPub(n.Pubkey()), input, sig) {
		return errInvalidSig
	}
	return nil
}

type hashFn func() hash.Hash

// ecdh creates a shared secret.
func ecdh(privkey *ecdsa.PrivateKey, pubkey *ecdsa.PublicKey) []byte {
	secX, secY := pubkey.ScalarMult(pubkey.X, pubkey.Y, privkey.D.Bytes())
	if secX == nil {
		return nil
	}
	sec := make([]byte, 33)
	sec[0] = 0x02 | byte(secY.Bit(0))
	math.ReadBits(secX, sec[1:])
	return sec
}

// deriveKeys creates the session keys.
func deriveKeys(hash hashFn, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, n1, n2 admnode.NodeID, challenge []byte) *session {
	var info = make([]byte, 0, len(n1)+len(n2))
	info = append(info, n1[:]...)
	info = append(info, n2[:]...)

	eph := ecdh(priv, pub)
	if eph == nil {
		return nil
	}
	kdf := hkdf.New(hash, eph, challenge, info)
	sec := session{writekey: make([]byte, aesKeySize), readkey: make([]byte, aesKeySize)}
	kdf.Read(sec.writekey)
	kdf.Read(sec.readkey)
	for i := range eph {
		eph[i] = 0
	}
	return &sec
}
