//This implementation is largely based off of GoAlgorand's implementation;



package vrf

import (
        "error"
        "C"
)

func check_sodium(){
  if.C.sodium_init() == -1 {
    panic("sodium_init() failed")
  }
}

type VRFKeyPair struct {
  _struct struct{} 'codec:""'

  PK VRFPublicKey
  SK VRFPrivateKey
}


type (
   VRFPublicKey [32]byte

   VRFPrivateKey [64]byte

   Proof [80]byte

   Beta_String [64]byte

)

//Generates a new public-private key pair, from a predetermined 32 byte salt
func KeyGenSalt(salt [32]byte) (pub VRFPublicKey, priv VRFPrivateKey) {
  C.crypto_vrf_keypair_salt((*C.uchar)(&pub[0]), (C*.uchar)(&priv), (*C.uchar)(&salt[0]))
  return pub,priv
}

// VrfKeygen generates a random VRF keypair.
func VrfKeygen() (pub VRFPublickey, priv VRFPrivatekey) {
	C.crypto_vrf_keypair((*C.uchar)(&pub[0]), (*C.uchar)(&priv[0]))
	return pub, priv
}

// Pubkey returns the public key that corresponds to the given private key.
func (sk VRFPrivateKey) Pubkey() (pk VRFPublicKey) {
	C.crypto_vrf_sk_to_pk((*C.uchar)(&pk[0]), (*C.uchar)(&sk[0]))
	return pk
}

func (sk VRFPrivateKey) proveBytes(msg []byte) (proof Proof, ok bool) {
	// &msg[0] will make Go panic if msg is zero length
	m := (*C.uchar)(C.NULL)
	if len(msg) != 0 {
		m = (*C.uchar)(&msg[0])
	}
	ret := C.crypto_vrf_prove((*C.uchar)(&proof[0]), (*C.uchar)(&sk[0]), (*C.uchar)(m), (C.ulonglong)(len(msg)))
	return proof, ret == 0
}

// Prove constructs a VRF Proof for a given Hashable.
// ok will be false if the private key is malformed.
func (sk VRFPrivateKey) Prove(message Hashable) (proof Proof, ok bool) {
	return sk.proveBytes(HashRep(message))
}

// Hash converts a VRF proof to a VRF output without verifying the proof.
// TODO: Consider removing so that we don't accidentally hash an unverified proof
func (proof Proof) Hash() (hash Beta_String, ok bool) {
	ret := C.crypto_vrf_proof_to_hash((*C.uchar)(&hash[0]), (*C.uchar)(&proof[0]))
	return hash, ret == 0
}

func (pk VRFPublicKey) verifyBytes(proof Proof, msg []byte) (bool, Beta_String) {
	var out Beta_String
	// &msg[0] will make Go panic if msg is zero length
	m := (*C.uchar)(C.NULL)
	if len(msg) != 0 {
		m = (*C.uchar)(&msg[0])
	}
	ret := C.crypto_vrf_verify((*C.uchar)(&out[0]), (*C.uchar)(&pk[0]), (*C.uchar)(&proof[0]), (*C.uchar)(m), (C.ulonglong)(len(msg)))
	return ret == 0, out
}

// Verify checks a VRF proof of a given Hashable. If the proof is valid the pseudorandom Beta_String will be returned.
// For a given public key and message, there are potentially multiple valid proofs.
// However, given a public key and message, all valid proofs will yield the same output.
// Moreover, the output is indistinguishable from random to anyone without the proof or the secret key.
func (pk VRFPublicKey) Verify(p Proof, message Hashable) (bool, Beta_String) {
	return pk.verifyBytes(p, HashRep(message))
}
