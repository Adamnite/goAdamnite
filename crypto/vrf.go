package crypto

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"io"

	"github.com/adamnite/go-adamnite/crypto/edwards25519"
	"github.com/adamnite/go-adamnite/crypto/extra25519"
	"golang.org/x/crypto/sha3"
)

const (
	PrivateKeySize   = 64
	PublicKeySize    = 32
	intermediateSize = 32
	ProofSize        = 32 + 32 + intermediateSize
)

type PrivateKey []byte
type PublicKey []byte

func GenerateVRFKey(rnd io.Reader) (sk PrivateKey, err error) {
	if rnd == nil {
		rnd = rand.Reader
	}

	sk = make([]byte, 64)
	_, err = io.ReadFull(rnd, sk[:32])
	if err != nil {
		return nil, err
	}

	x, _ := sk.expandSecret()

	var pkP edwards25519.ExtendedGroupElement
	edwards25519.GeScalarMultBase(&pkP, x)

	var pkBytes [PublicKeySize]byte
	pkP.ToBytes(&pkBytes)

	copy(sk[32:], pkBytes[:])
	return sk, nil
}

// Prove returns the vrf value and a proof.
// Verify(seed, vrf, proof) == true If the vrf value is the same as returned by Compute(seed).
func (sk PrivateKey) Prove(seed []byte) (vrf, proof []byte) {
	x, skhr := sk.expandSecret()

	var (
		sH, rH                                 [64]byte
		r, s, minusS, t, gB, grB, hrB, hxB, hB [32]byte
		ii, gr, hr                             edwards25519.ExtendedGroupElement
	)

	h := hashToCurve(seed)

	h.ToBytes(&hB)
	edwards25519.GeScalarMult(&ii, x, h)
	ii.ToBytes(&hxB)

	// use hash of private-, public-key and msg as randomness source:
	hash := sha3.NewShake256()
	hash.Write(skhr[:])
	hash.Write(sk[32:]) // public key, as in ed25519
	hash.Write(seed)
	hash.Read(rH[:])
	hash.Reset()
	edwards25519.ScReduce(&r, &rH)

	edwards25519.GeScalarMultBase(&gr, &r)
	edwards25519.GeScalarMult(&hr, &r, h)
	gr.ToBytes(&grB)
	hr.ToBytes(&hrB)
	gB = edwards25519.BaseBytes

	// H2(g, h, g^x, h^x, g^r, h^r, seed)
	hash.Write(gB[:])
	hash.Write(hB[:])
	hash.Write(sk[32:]) // ed25519 public-key
	hash.Write(hxB[:])
	hash.Write(grB[:])
	hash.Write(hrB[:])
	hash.Write(seed)
	hash.Read(sH[:])
	hash.Reset()
	edwards25519.ScReduce(&s, &sH)

	edwards25519.ScNeg(&minusS, &s)
	edwards25519.ScMulAdd(&t, x, &minusS, &r)

	proof = make([]byte, ProofSize)
	copy(proof[:32], s[:])
	copy(proof[32:64], t[:])
	copy(proof[64:96], hxB[:])

	hash.Write(hxB[:])
	hash.Write(seed)
	vrf = make([]byte, 32)
	hash.Read(vrf[:])
	return vrf, proof
}

// Verify returns true iff vrf=Compute(seed) for the sk that
// corresponds to pk.
func (pkBytes PublicKey) Verify(seed, vrfBytes, proof []byte) bool {
	if len(proof) != ProofSize || len(vrfBytes) != 32 || len(pkBytes) != PublicKeySize {
		return false
	}
	var pk, s, sRef, t, vrf, hxB, hB, gB, ABytes, BBytes [32]byte
	copy(vrf[:], vrfBytes)
	copy(pk[:], pkBytes[:])
	copy(s[:32], proof[:32])
	copy(t[:32], proof[32:64])
	copy(hxB[:], proof[64:96])

	hash := sha3.NewShake256()
	hash.Write(hxB[:]) // const length
	hash.Write(seed)
	var hCheck [32]byte
	hash.Read(hCheck[:])
	if !bytes.Equal(hCheck[:], vrf[:]) {
		return false
	}
	hash.Reset()

	var P, B, ii, iic edwards25519.ExtendedGroupElement
	var A, hmtP, iicP edwards25519.ProjectiveGroupElement
	if !P.FromBytesBaseGroup(&pk) {
		return false
	}
	if !ii.FromBytesBaseGroup(&hxB) {
		return false
	}
	edwards25519.GeDoubleScalarMultVartime(&A, &s, &P, &t)
	A.ToBytes(&ABytes)
	gB = edwards25519.BaseBytes

	h := hashToCurve(seed) // h = H1(seed)
	h.ToBytes(&hB)
	edwards25519.GeDoubleScalarMultVartime(&hmtP, &t, h, &[32]byte{})
	edwards25519.GeDoubleScalarMultVartime(&iicP, &s, &ii, &[32]byte{})
	iicP.ToExtended(&iic)
	hmtP.ToExtended(&B)
	edwards25519.GeAdd(&B, &B, &iic)
	B.ToBytes(&BBytes)

	var sH [64]byte
	// sRef = H2(g, h, g^x, v, g^t·G^s,H1(seed)^t·v^s, seed), with v=H1(seed)^x=h^x
	hash.Write(gB[:])
	hash.Write(hB[:])
	hash.Write(pkBytes)
	hash.Write(hxB[:])
	hash.Write(ABytes[:]) // const length (g^t*G^s)
	hash.Write(BBytes[:]) // const length (H1(seed)^t*v^s)
	hash.Write(seed)
	hash.Read(sH[:])

	edwards25519.ScReduce(&sRef, &sH)
	return sRef == s
}

func (sk PrivateKey) expandSecret() (x, skhr *[32]byte) {
	x, skhr = new([32]byte), new([32]byte)

	hash := sha3.NewShake256()
	hash.Write(sk[:32])
	hash.Read(x[:])
	hash.Read(skhr[:])

	x[0] &= 248
	x[31] &= 127
	x[31] |= 64
	return x, skhr
}

func (sk PrivateKey) Public() (PublicKey, bool) {
	pk, ok := ed25519.PrivateKey(sk).Public().(ed25519.PublicKey)
	return PublicKey(pk), ok
}

func hashToCurve(seed []byte) *edwards25519.ExtendedGroupElement {
	// H(n) = (f(h(n))^8)
	var hmb [32]byte
	sha3.ShakeSum256(hmb[:], seed)
	var hm edwards25519.ExtendedGroupElement
	extra25519.HashToEdwards(&hm, &hmb)
	edwards25519.GeDouble(&hm, &hm)
	edwards25519.GeDouble(&hm, &hm)
	edwards25519.GeDouble(&hm, &hm)
	return &hm
}

// Compute generates the vrf value for the byte slice seed using the
// underlying private key sk.
func (sk PrivateKey) Compute(seed []byte) []byte {
	x, _ := sk.expandSecret()
	var ii edwards25519.ExtendedGroupElement
	var iiB [32]byte
	edwards25519.GeScalarMult(&ii, x, hashToCurve(seed))
	ii.ToBytes(&iiB)

	hash := sha3.NewShake256()
	hash.Write(iiB[:]) // const length: Size
	hash.Write(seed)
	var vrValue [32]byte
	hash.Read(vrValue[:])
	return vrValue[:]
}
