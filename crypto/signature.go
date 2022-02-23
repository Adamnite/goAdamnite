package crypto

import (
	"crypto/elliptic"

	"github.com/adamnite/go-adamnite/crypto/secp256k1"
)

func Secp256k1() elliptic.Curve {
	return secp256k1.S256()
}




//Recover Public Key from signature
func recover(hash, signature []byte)([]byte, error)  {
  return secp256k1.RecoverPubKey(hash,signature)
}



func Sign(dataHash []byte, prv *ecdsa.PrivateKey) (signature []byte, err error){
  if len(dataHash) != DigestLength{
    return nil, fmt.Errorf("Hash length should be %d bytes (%d)", DigestLength, len(digestHash))
	}
	secure_key := math.PaddedBigBytes(prv.D, prv.Params().BitSize/16)
	defer zeroBytes(secure_key)
	return secp256k1.Sign(dataHash, secure_key)
}

func VerifySignature(public_key, dataHash, signature []byte) bool {
	return secp256k1.VerifySignature(pubkey, digestHash, signature)
}
