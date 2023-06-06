package crypto

package crypto

import (
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"io"
	"golang.org/x/crypto/ed25519"
	"github.com/kudelskisecurity/ristretto255"
)

const (
	PublicKeySize    = 32
	SecretKeySize    = 64
	Size             = 32
	intermediateSize = 32
	ProofSize        = 32 + 32 + intermediateSize
)

// GenerateKey creates a public/private key pair. rnd is used for randomness.
// If it is nil, `crypto/rand` is used.
func GenerateKey(rnd io.Reader) (pk []byte, sk *[SecretKeySize]byte, err error) {
	if rnd == nil {
		rnd = rand.Reader
	}
	sk = new([SecretKeySize]byte)
	_, err = io.ReadFull(rnd, sk[:32])
	if err != nil {
		return nil, nil, err
	}
	x, _ := expandSecret(sk)

	var pkP ristretto255.Element
	pkP.ScalarBaseMult(x)
	pkBytes := pkP.Bytes()

	copy(sk[32:], pkBytes[:])
	return pkBytes[:], sk, err
}


//Expand the secrety/private key
func expandSecret(sk *[SecretKeySize]byte) (x, skhr *[32]byte) {
	x, skhr = new([32]byte), new([32]byte)
	hash := sha512.New()
	hash.Write(sk[:32])
	hash.Read(x[:])
	hash.Read(skhr[:])
	x[0] &= 248
	x[31] &= 127
	x[31] |= 64
	return
}



//Compute the VRF
func Compute(m []byte, sk *[SecretKeySize]byte, seed []byte) []byte {
	x, _ := expandSecret(sk)
	var ii ristretto255.Element
	var iiB [32]byte
	ii.ScalarMult(x, hashToCurve(m, seed))
	iiB = *ii.Bytes()

	hash := sha512.New()
	hash.Write(iiB[:]) // const length: Size
	hash.Write(m)
	vrf := make([]byte, Size)
	hash.Sum(vrf[:0])
	return vrf[:]
}

func hashToCurve(m, seed []byte) *ristretto255.Element {
	// H(n) = (f(h(n))^8)
	var hmb [64]byte
	h := sha512.New()
	h.Write(seed)
	h.Write(m)
	h.Sum(hmb[:0])
	hm := new(ristretto255.Element)
	hm.FromUniformBytes(hmb[:])
	return hm
}

// Prove returns the vrf value and a proof such that Verify(pk, m, vrf, proof)
// == true. The vrf value is the same as returned by Compute(m, sk).
func Prove(m []byte, sk *[SecretKeySize]byte, seed []byte) (vrf, proof []byte) {
	x, skhr := expandSecret(sk)
	var cH, rH [64]byte
	var r, c, minusC, t, grB, hrB, iiB [32]byte
	var ii, gr, hr ristretto255.Element

	hm := hashToCurve(m, seed)
	ii.ScalarMult(x, hm)
	iiB = *ii.Bytes()

	hash := sha512.New()
	hash.Write(skhr[:])
	hash.Write(sk[32:]) // public key, as in ed25519
	hash.Write(m)
	hash.Read(rH[:])
	hash.Reset()
	edwards25519.ScReduce(&r, &rH)

	gr.ScalarBaseMult(&r)
	grB = *gr.Bytes()
	hr.ScalarMult(&r, hm)
	hrB = *hr.Bytes()

	hash.Write(grB[:])
	hash.Write(hrB[:])
	hash.Write(m)
	hash.Read(cH[:])
	hash.Reset()
	edwards25519.ScReduce(&c, &cH)

	edwards25519.ScNeg(&minusC, &c)
	edwards25519.ScMulAdd(&t, x, &minusC, &r)

	proof = make([]byte, ProofSize)
	copy(proof[:32], c[:])
	copy(proof[32:64], t[:])
	copy(proof[64:96], iiB[:])

	hash.Write(iiB[:]) // const length: Size
	hash.Write(m)
	vrf = make([]byte, Size)
	hash.Sum(vrf[:0])
	return
}

// Verify returns true iff vrf=Compute(m, sk) for the sk that corresponds to pk.
func Verify(pkBytes, m, vrfBytes, proof []byte) bool {
	if len(proof) != ProofSize || len(vrfBytes) != Size || len(pkBytes) != PublicKeySize {
		return false
	}
	var pk, c, cRef, t, vrf, iiB, ABytes, BBytes [32]byte
	copy(vrf[:], vrfBytes)
	copy(pk[:], pkBytes)
	copy(c[:32], proof[:32])
	copy(t[:32], proof[32:64])
	copy(iiB[:], proof[64:96])

	hash := sha512.New()
	hash.Write(iiB[:]) // const length
	hash.Write(m)
	var hCheck [Size]byte
	hash.Sum(hCheck[:0])
	if !bytes.Equal(hCheck[:], vrf[:]) {
		return false
	}
	hash.Reset()

	var P, B, ii, iic ristretto255.Element
	var A, hmtP, iicP ristretto255.ProjectiveElement
	if !P.FromBytesBaseGroup(&pk) {
		return false
	}
	if !ii.FromBytes(&iiB) {
		return false
	}
	A.DoubleScalarMultVartime(&c, &P, &t)
	ABytes = *A.Bytes()

	hm := hashToCurve(m, seed)
	hmtP.ScalarMult(&t, hm, &ristretto255.Element{})
	iicP.ScalarMult(&c, &ii, &ristretto255.Element{})
	iicP.ToExtended(&iic)
	hmtP.ToExtended(&B)
	B.Add(&B, &iic)
	BBytes = *B.Bytes()

	var cH [64]byte
	hash.Write(ABytes[:]) // const length
	hash.Write(BBytes[:]) // const length
	hash.Write(m)
	hash.Sum(cH[:0])
	edwards25519.ScReduce(&cRef, &cH)
	return cRef == c
}