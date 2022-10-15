package types

import (
	"bytes"
	"errors"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrorCryptoSignature    = errors.New("Invalid Elliptic Curve Values")
	ErrorEmptyTX            = errors.New("Empty Transaction Data")
	ErrorLowFee             = errors.New("Fee lower than minimum fee dicated by network")
	ErrorInvalidTransaction = errors.New("Transaction Balance is too large")
)

type TxType int8

const (
	VOTE_TX TxType = iota
	NORMAL_TX
	CONTRACT_TX
)

// Transaction is an Adamnite transaction.
type Transaction struct {
	InnerData Transaction_Data
	timestamp time.Time
	//Cache Values, similar to GoETH
	From atomic.Value
	hash atomic.Value
	size atomic.Value
}

func CreateTx(InnerData Transaction_Data) *Transaction {
	transaction := new(Transaction)
	transaction.Decode(InnerData.copy(), 0) //We will need to some functions for decoding and encoding data types
	//This can be Serialization or our own implementation.
	return transaction
}

// TxData is the underlying data of a transaction.
type Transaction_Data interface {
	// copy creates a deep copy and initializes all fields
	txtype() TxType
	copy() Transaction_Data
	chain_TYPE() *big.Int
	to() *common.Address
	amount() *big.Int
	message() []byte
	message_size() *big.Int //Size of message in bytes
	ATE_MAX() uint64        //Transaction Fee Limit that the user encoded
	ATE_price() *big.Int    //Actual Transaction Fee
	nonce() uint64

	rawSignature() (v, r, s *big.Int)
	setSignatue(chain_TYPE, v, r, s *big.Int)
}

type Transactions []*Transaction

func (s Transactions) Len() int { return len(s) }

func (s Transactions) EncodeIndex(i int, w *bytes.Buffer) {
	tx := s[i]
	tx.encodeTyped(w)
}

//// TODO: Implement structure for creating a new transaction, basic message transactions (will do this weekend),
//proper encoding and decoding of transactions, fetch operations (the fetching of various transaction releated data),
//and some stack operations such as shift.

func NewTx(inner Transaction_Data) *Transaction {
	tx := new(Transaction)
	tx.setDecoded(inner.copy(), 0)
	return tx
}

func (tx *Transaction) Decode(s Transaction_Data, _ int) {

}

// setDecoded sets the inner transaction and size after decoding.
func (tx *Transaction) setDecoded(inner Transaction_Data, size int) {
	tx.InnerData = inner
	tx.timestamp = time.Now()
	if size > 0 {
		tx.size.Store(common.StorageSize(size))
	}
}

func (tx *Transaction) Type() TxType {
	return tx.InnerData.txtype()
}

func (tx *Transaction) RawSignature() (v, r, s *big.Int) {
	return tx.InnerData.rawSignature()
}

// WithSignature returns a new transaction with the given signature.
func (tx *Transaction) WithSignature(signer Signer, signature []byte) (*Transaction, error) {
	r, s, v, err := signer.SignatureValues(tx, signature)
	if err != nil {
		return nil, err
	}

	cpy := tx.InnerData.copy()
	cpy.setSignatue(signer.ChainType(), v, r, s)
	return &Transaction{InnerData: cpy, timestamp: tx.timestamp}, nil
}

// Nonce returns the sender account nonce of the transaction.
func (tx *Transaction) Nonce() uint64 { return tx.InnerData.nonce() }

func (tx *Transaction) ATEMax() uint64 { return tx.InnerData.ATE_MAX() }

func (tx *Transaction) ATEPrice() *big.Int { return tx.InnerData.ATE_price() }

func (tx *Transaction) Amount() *big.Int { return tx.InnerData.amount() }

func (tx *Transaction) Cost() *big.Int {
	total := new(big.Int).Mul(new(big.Int).SetUint64(tx.ATEMax()), tx.ATEPrice())
	total.Add(total, tx.Amount())
	return total
}

func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	msgpack.NewEncoder(&c).Encode(&tx.InnerData)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}

	var h common.Hash
	h = prefixedSerializationHash(byte(tx.Type()), tx.InnerData)
	tx.hash.Store(h)
	return h
}

// encodeTyped writes the canonical encoding of a typed transaction to w.
func (tx *Transaction) encodeTyped(w *bytes.Buffer) error {
	w.WriteByte(byte(tx.Type()))
	return msgpack.NewEncoder(w).Encode(tx.InnerData)
}

type TxByNonce Transactions

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].Nonce() < s[j].Nonce() }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (tx *Transaction) ATEPriceCmp(other *Transaction) int {
	return tx.InnerData.ATE_price().Cmp(other.InnerData.ATE_price())
}

func (tx *Transaction) ATEPriceIntCmp(other *big.Int) int {
	return tx.InnerData.ATE_price().Cmp(other)
}
