package types

import (
	"bytes"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/vmihailenco/msgpack/v5"
)

type TxType int8

const (
	VOTE_TX TxType = iota
	VOTE_POH_TX
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
	setSignature(chain_TYPE, v, r, s *big.Int)
}

type Transactions []*Transaction

func (s Transactions) Len() int { return len(s) }

func (s Transactions) EncodeIndex(i int, w *bytes.Buffer) {
	tx := s[i]
	tx.encodeTyped(w)
}

//// TODO: Implement structure for creating a new transaction, basic message transactions (will do this weekend),
//proper encoding and decoding of transactions, fetch operations (the fetching of various transaction related data),
//and some stack operations such as shift.

func NewTx(inner Transaction_Data) *Transaction {
	tx := new(Transaction)
	tx.setDecoded(inner.copy(), 0)
	return tx
}

func (tx *Transaction) To() *common.Address {
	return tx.InnerData.to()
}

func (tx *Transaction) Decode(s Transaction_Data, _ int) {

}

type Message struct {
	to         *common.Address
	from       common.Address
	nonce      uint64
	amount     *big.Int
	gasLimit   uint64
	gasPrice   *big.Int
	data       []byte
	checkNonce bool
}

func (msg Message) From() common.Address {
	return msg.from
}
func (msg Message) To() *common.Address {
	return msg.to
}
func (msg Message) AtePrice() *big.Int {
	return msg.gasPrice
}
func (msg Message) Ate() uint64 {
	return msg.gasLimit
}
func (msg Message) Value() *big.Int {
	return msg.amount
}
func (msg Message) Nonce() uint64 {
	return msg.nonce
}
func (msg Message) CheckNonce() bool {
	return msg.checkNonce
}
func (msg Message) Data() []byte {
	return msg.data
}

func (tx *Transaction) AsMessage(s Signer) (Message, error) {
	msg := Message{
		nonce: tx.InnerData.nonce(),

		gasPrice:   tx.ATEPrice(),
		to:         tx.InnerData.to(),
		amount:     tx.InnerData.amount(),
		data:       tx.InnerData.message(),
		checkNonce: true,
	}

	var err error
	msg.from, err = Sender(s, tx)
	return msg, err
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
	cpy.setSignature(signer.ChainType(), v, r, s)
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
