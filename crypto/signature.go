package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	"github.com/adamnite/goadamnite/common"
	"github.com/adamnite/goadamnite/crypto/secp256k1"
)

//Use the elliptic curve methods to sign data.
//Different helper functions such as recovring the public key 
//from the signature are also included.

func SignData(data []byte, privKeyBytes []byte) ([]byte, error) {
    privKey, _ := secp256k1.PrivKeyFromBytes(privKeyBytes)
    sig, err := secp256k1.Sign(data, privKey)
    if err != nil {
        return nil, fmt.Errorf("failed to sign data: %v", err)
    }
    return sig, nil
}

//Recover Public Key from signature
func Recover(hash []byte, R, S, Vb *big.Int) (*ecdsa.PublicKey, error) {
    curve := secp256k1()

    V := byte(Vb.Uint64() - 27)
    if V != 0 && V != 1 {
        return nil, fmt.Errorf("invalid recovery ID")
    }

    // Compute the x-coordinate of the public key
    x, y := elliptic.Unmarshal(curve, pubkeyBytes[1:])
    if x == nil {
        return nil, fmt.Errorf("invalid public key")
    }

    // Compute the y-coordinate of the public key
    isYEven := y.Bit(0) == 0
    if isYEven != (V == 0) {
        y = curve.Params().P.Sub(y, new(big.Int).Set(y))
    }

    // Create and return the public key
    return secp256k1.PublicKey{curve, x, y}, nil
}

func VerifySignature(pubKeyBytes, hash, sig []byte) (bool, error) {
    pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
    if err != nil {
        return false, fmt.Errorf("failed to parse public key: %v", err)
    }

    isValid := secp256k1.VerifySignature(pubKey, hash, sig[:len(sig)-1])
    return isValid, nil
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func CompressPublicKey(pubKeyBytes []byte) ([]byte, error) {
    pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
    if err != nil {
        return nil, fmt.Errorf("failed to parse public key: %v", err)
    }

    compressedPubKey := secp256k1.CompressPubkey(pubKey)

    // The compressed public key should be 33 bytes long, with the first byte indicating
    // whether the y-coordinate is even or odd (0x02 for even, 0x03 for odd).
    if len(compressedPubKey) != 33 {
        return nil, fmt.Errorf("unexpected compressed public key length")
    }

    if compressedPubKey[0] != 0x02 && compressedPubKey[0] != 0x03 {
        return nil, fmt.Errorf("unexpected compressed public key format")
    }

    return compressedPubKey, nil
}

func DecompressPubkey(pubkey []byte) ([]byte, error) {
	if len(pubkey) != 33 {
		return nil, fmt.Errorf("invalid compressed public key length")
	}

	// Decompress the public key
	uncompressedPubKey, err := secp256k1.DecompressPubkey(pubkey)
	if err != nil {
		return nil, err
	}
	return uncompressedPubKey, nil
}


type OneTimeSignature struct {
	pubKey     ecdsa.PublicKey
	privKey    ecdsa.PrivateKey
	counter    uint32
	maxCounter uint32
}

type OneTimeSignatureEphemeral struct {
	privKey ecdsa.PrivateKey
}

func NewOneTimeSignature(maxCounter uint32) (*OneTimeSignature, error) {
	curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}

	return &OneTimeSignature{
		pubKey:     privKey.PublicKey,
		privKey:    *privKey,
		counter:    0,
		maxCounter: maxCounter,
	}, nil
}

func (ots *OneTimeSignature) Sign(msg []byte) (*big.Int, *big.Int, error) {
	if ots.counter >= ots.maxCounter {
		return nil, nil, errors.New("max signature limit reached")
	}

	ephKey, err := NewOneTimeSignatureEphemeral()
	if err != nil {
		return nil, nil, err
	}

	h := sha256.Sum256(msg)
	r, s, err := ecdsa.Sign(rand.Reader, &ephKey.privKey, h[:])
	if err != nil {
		return nil, nil, err
	}

	// Calculate one-time public key
	Rx, Ry := curve.ScalarBaseMult(r.Bytes())
	Px, Py := curve.ScalarMult(ots.pubKey.X, ots.pubKey.Y, s.Bytes())
	Qx, Qy := curve.Add(Rx, Ry, Px, Py)
	if !curve.IsOnCurve(Qx, Qy) {
		return nil, nil, errors.New("invalid signature")
	}

	ots.counter++

	return Qx, Qy, nil
}

func (ots *OneTimeSignature) Verify(msg []byte, Qx, Qy *big.Int) error {
	if ots.counter >= ots.maxCounter {
		return errors.New("max signature limit reached")
	}

	h := sha256.Sum256(msg)

	if !ecdsa.Verify(&ots.pubKey, h[:], Qx, Qy) {
		return errors.New("invalid signature")
	}

	ots.counter++

	return nil
}

func NewOneTimeSignatureEphemeral() (*OneTimeSignatureEphemeral, error) {
	curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}

	return &OneTimeSignatureEphemeral{
		privKey: *privKey,
	}, nil
}



// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return secp256k1.S256()
}