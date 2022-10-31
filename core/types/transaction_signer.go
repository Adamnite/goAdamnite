package types

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/serialization"
)

var (
	ErrTxTypeNotSupported = errors.New("transaction type not supported")
	ErrInvalidSig         = errors.New("Invalid signature")
)

type Signer interface {
	// Sender returns the sender address of the transaction.
	Sender(tx *Transaction) (common.Address, error)

	// SignatureValues returns the R, S, V values
	SignatureValues(tx *Transaction, signature []byte) (r, s, v *big.Int, err error)

	// ChainType returns the current chain type
	ChainType() *big.Int

	// Hash returns the signature hash
	Hash(tx *Transaction) common.Hash
}

type AdamniteSigner struct{}

func (as AdamniteSigner) Sender(tx *Transaction) (common.Address, error) {
	if tx.Type() != VOTE_TX && tx.Type() != NORMAL_TX {
		return common.Address{}, ErrTxTypeNotSupported
	}
	v, r, s := tx.RawSignature()
	return recoverPlain(as.Hash(tx), r, s, v)
}

func (as AdamniteSigner) Hash(tx *Transaction) common.Hash {
	serial := serialization.Serialize(tx.Nonce())
	bytes := crypto.Sha512(serial)
	hash := common.Hash{}
	copy(hash[:], bytes)
	return hash
}

func (as AdamniteSigner) ChainType() *big.Int {
	return nil
}

func (as AdamniteSigner) SignatureValues(tx *Transaction, signature []byte) (r, s, v *big.Int, err error) {
	if tx.Type() != VOTE_TX && tx.Type() != NORMAL_TX {
		return nil, nil, nil, ErrTxTypeNotSupported
	}
	r, s, v = decodeSignature(signature)
	return r, s, v, nil
}

func decodeSignature(sig []byte) (r, s, v *big.Int) {
	if len(sig) != crypto.SignatureLength {
		panic(fmt.Sprintf("wrong size for signature: got %d, want %d", len(sig), crypto.SignatureLength))
	}
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64] + 27})
	return r, s, v
}

func recoverPlain(sighash common.Hash, R, S, Vb *big.Int) (common.Address, error) {
	if Vb.BitLen() > 8 {
		return common.Address{}, ErrInvalidSig
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S) {
		return common.Address{}, ErrInvalidSig
	}
	// encode the signature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, crypto.SignatureLength)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the signature
	pub, err := crypto.Recover(sighash[:], sig)
	if err != nil {
		return common.Address{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return common.Address{}, errors.New("invalid public key")
	}
	var addr common.Address
	copy(addr[:], crypto.Ripemd160Hash(crypto.Sha512(pub[1:])))
	return addr, nil
}

func SignTransaction(tx *Transaction, s Signer, prv *ecdsa.PrivateKey) (*Transaction, error) {
	hash := s.Hash(tx)
	signature, err := crypto.Sign(hash[:], prv)
	if err != nil {
		return nil, err
	}

	return tx.WithSignature(s, signature)
}

func Sender(signer Signer, tx *Transaction) (common.Address, error) {
	addr, err := signer.Sender(tx)
	if err != nil {
		return common.Address{}, err
	}

	return addr, nil
}
