package types

import (
	"bytes"
	"container/heap"
	"errors"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/bytes"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrorCryptoSignature    = errors.New("Invalid Elliptic Curve Values")
	ErrorEmptyTX            = errors.New("Empty Transaction Data")
	ErrorLowFee             = errors.New("Fee lower than minimum fee dictated by network")
	ErrorInvalidTransaction = errors.New("Transaction Balance is too large")
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
	to() *bytes.Address
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

func (tx *Transaction) To() *bytes.Address {
	return tx.InnerData.to()
}

func (tx *Transaction) Decode(s Transaction_Data, _ int) {

}

type Message struct {
	to         *bytes.Address
	from       bytes.Address
	nonce      uint64
	amount     *big.Int
	gasLimit   uint64
	gasPrice   *big.Int
	data       []byte
	checkNonce bool
}

func (msg Message) From() bytes.Address {
	return msg.from
}
func (msg Message) To() *bytes.Address {
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

func (tx *Transaction) Hash() bytes.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(bytes.Hash)
	}

	var h bytes.Hash
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

type TxByPrice Transactions

func (s TxByPrice) Len() int { return len(s) }
func (s TxByPrice) Less(i, j int) bool {
	// If the prices are equal, use the time the transaction was first seen for
	// deterministic sorting
	cmp := s[i].ATEPrice().Cmp(s[j].ATEPrice())
	if cmp == 0 {
		return s[i].timestamp.Before(s[j].timestamp)
	}
	return cmp > 0
}
func (s TxByPrice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s *TxByPrice) Push(x interface{}) {
	*s = append(*s, x.(*Transaction))
}

func (s *TxByPrice) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type TransactionsByPriceAndNonce struct {
	txs    map[bytes.Address]Transactions //List of transaction records currently sorted by account
	heads  TxByPrice                       // next transaction for each unique account (price heap)
	signer Signer                          //The signer of the transaction set
}

// newTransactionByPriceAndOnce creates a retrieving
// Sort the trades by price in a non-cash way.
//
// Note that the input map is re-owned, so the caller should no longer interact with
// if after provided to the constructor.
func NewTransactionsByPriceAndNonce(signer Signer, txs map[bytes.Address]Transactions) *TransactionsByPriceAndNonce {
	// Initialize a price and received time based heap with the head transactions
	heads := make(TxByPrice, 0, len(txs))
	for from, accTxs := range txs {
		// Ensure the sender address is from the signer
		if acc, _ := Sender(signer, accTxs[0]); acc != from {
			delete(txs, from)
			continue
		}
		heads = append(heads, accTxs[0])
		txs[from] = accTxs[1:]
	}
	heap.Init(&heads)

	//Assemble and return the transaction set
	return &TransactionsByPriceAndNonce{
		txs:    txs,
		heads:  heads,
		signer: signer,
	}
}

// Peek returns the next transaction by price.
func (t *TransactionsByPriceAndNonce) Peek() *Transaction {
	if len(t.heads) == 0 {
		return nil
	}
	return t.heads[0]
}
