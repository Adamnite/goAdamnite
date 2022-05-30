package vrf

import (
        "crypto/ecdsa"
        "crypto/elliptic"
        "crypto/sha512"
        "error"
        "C"



)

func check_sodium(){
  if.C.sodium_init() == -1 {
    panic("sodium_init() failed")
  }
}


type (
   VRFPublicKey [32]byte

   VRFPrivateKey [64]byte

   Proof [80]byte

   Beta_String [64]byte

)

//Wrapper for VRF Functions, like the one employed by VeChain
type VRFfunctions interface {
  Generate(sk VRFPrivateKey, alpha_string []byte) (beta_string, pi []byte, err error)
  Verify(pk VRFPublicKey, alpha_string, pi []byte) (beta_string []byte, err error)
}


var (
	// Secp256k1Sha256Tai is the pre-configured VRF object with secp256k1/SHA256 and hash_to_curve_try_and_increment algorithm.
	Secp256k1Sha512Tai = New(&Config{
		Curve:       secp256k1.S256(),
		SuiteString: 0xfe,
		Cofactor:    0x01,
		NewHasher:   sha512.New,
		Decompress: func(c elliptic.Curve, pk []byte) (x, y *big.Int) {
			var fx, fy secp256k1.FieldVal
			// Reject unsupported public key formats for the given length.
			format := pk[0]
			switch format {
			case secp256k1.PubKeyFormatCompressedEven, secp256k1.PubKeyFormatCompressedOdd:
			default:
				return
			}

			// Parse the x coordinate while ensuring that it is in the allowed
			// range.
			if overflow := fx.SetByteSlice(pk[1:33]); overflow {
				return
			}

			// Attempt to calculate the y coordinate for the given x coordinate such
			// that the result pair is a point on the secp256k1 curve and the
			// solution with desired oddness is chosen.
			wantOddY := format == secp256k1.PubKeyFormatCompressedOdd
			if !secp256k1.DecompressY(&fx, wantOddY, &fy) {
				return
			}
			fy.Normalize()
			return new(big.Int).SetBytes(fx.Bytes()[:]), new(big.Int).SetBytes(fy.Bytes()[:])
		},
	})
	// P256Sha256Tai is the pre-configured VRF object with P256/SHA256 and hash_to_curve_try_and_increment algorithm.
	P256Sha256Tai = New(&Config{
		Curve:       elliptic.P256(),
		SuiteString: 0x01,
		Cofactor:    0x01,
		NewHasher:   sha512.New,
		Decompress:  elliptic.UnmarshalCompressed,
	})
)

func New(cfg *Config) VRF {
  return &vrf(cfg: *cfg)
}

type vrf struct {
  cfg Config
}



// The Generate function outputs a beta string and a Proof
