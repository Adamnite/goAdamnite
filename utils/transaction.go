package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

type TransactionVersionType int8

const (
	// add more transaction versions as we develop our blockchain
	TRANSACTION_V0 TransactionVersionType = iota
)

type TransactionActionType = int8

const (
	Transaction_Basic TransactionActionType = iota
	Transaction_VM_Call
	Transaction_VM_NewContract
)

type TransactionType interface {
	FromAddress() common.Address
	GetType() TransactionActionType
	Hash() common.Hash
	VerifySignature() (bool, error)
	Equal(TransactionType) bool
	GetTime() time.Time
	GetSignature() []byte
	GetAmount() *big.Int
}

type BaseTransaction struct {
	Version         TransactionVersionType
	From            *accounts.Account
	TransactionType TransactionActionType
	To              common.Address
	Amount          *big.Int
	GasLimit        *big.Int
	Time            time.Time
	Signature       []byte
}

func NewBaseTransaction(sender *accounts.Account, to common.Address, value *big.Int, gasLimit *big.Int) (*BaseTransaction, error) {
	if value.Sign() == -1 {
		return nil, ErrNegativeAmount
	}
	t := BaseTransaction{
		Version:  TRANSACTION_V0,
		From:     sender,
		To:       to,
		Time:     time.Now().UTC(),
		Amount:   value,
		GasLimit: gasLimit.Abs(gasLimit), //sanitizing the gas limit passed.
	}
	sig, err := sender.Sign(t)
	t.Signature = sig
	return &t, err
}
func (t BaseTransaction) GetType() TransactionActionType {
	return t.TransactionType
}
func (t BaseTransaction) FromAddress() common.Address {
	return t.From.GetAddress()
}
func (t BaseTransaction) GetTime() time.Time {
	return t.Time
}
func (t BaseTransaction) GetSignature() []byte {
	return t.Signature
}
func (t BaseTransaction) GetAmount() *big.Int {
	return t.Amount
}

// does not use the time in the hash value, as that is used for the random source value.
func (t BaseTransaction) Hash() common.Hash {
	data := append(t.From.PublicKey, t.To.Bytes()...)
	data = append(data, t.Amount.Bytes()...)

	data = append(data, binary.LittleEndian.AppendUint64([]byte{}, uint64(t.Time.UnixMilli()))...)

	return common.BytesToHash(crypto.Sha512(data))
}

// Verify that the signature used in the transaction is correct
func (t BaseTransaction) VerifySignature() (ok bool, err error) {
	if t.Amount.Sign() != 1 {
		return false, ErrNegativeAmount
	}
	return t.From.Verify(t, t.Signature), nil
}

// Test equality between two transactions
func (a BaseTransaction) Equal(other TransactionType) bool {
	if a.GetType() != other.GetType() {
		return false
	}
	b := other.(BaseTransaction)
	return bytes.Equal(a.From.PublicKey, b.From.PublicKey) &&
		bytes.Equal(a.To.Bytes(), b.To.Bytes()) &&
		a.Amount.Cmp(b.Amount) == 0 &&
		bytes.Equal(a.Signature, b.Signature)
}

type VMCallTransaction struct {
	BaseTransaction
	VMInteractions *RuntimeChanges //an optional value that would make this transaction require chamber B validation
	RunnerHash     []byte          //the running hash that was used as the start of this VM.
}

// takes a transaction and changes it to be a VM interacting transaction
func NewVMTransactionFrom(signer *accounts.Account, buildOn *BaseTransaction, callParams []byte) (*VMCallTransaction, error) {
	vmCall := RuntimeChanges{
		Caller:           signer.GetAddress(),
		CallTime:         buildOn.Time,
		ContractCalled:   buildOn.To,
		ParametersPassed: callParams,
		GasLimit:         buildOn.GasLimit.Uint64(),
	}
	vmTransaction := VMCallTransaction{
		BaseTransaction: *buildOn,
		VMInteractions:  &vmCall,
	}
	vmTransaction.TransactionType = Transaction_VM_Call
	sig, err := signer.Sign(vmTransaction)
	vmTransaction.Signature = sig
	return &vmTransaction, err
}
func (vmt VMCallTransaction) Hash() common.Hash {
	data := append(vmt.From.PublicKey, vmt.To.Bytes()...)
	data = append(data, vmt.Amount.Bytes()...)

	data = append(data, binary.LittleEndian.AppendUint64([]byte{}, uint64(vmt.Time.UnixMilli()))...)
	data = append(data, vmt.VMInteractions.Hash().Bytes()...)

	return common.BytesToHash(crypto.Sha512(data))
}

type NewContractTransaction struct {
	Version         TransactionVersionType
	From            *accounts.Account
	TransactionType TransactionActionType
	Amount          *big.Int //still need to send some amount of funds to the contract when you make it.
	GasLimit        *big.Int
	Time            time.Time
	Signature       []byte
}

func MakeNewContractTransaction() (*NewContractTransaction, error) {
	nct := NewContractTransaction{}
	return &nct, nil
}
func (nct NewContractTransaction) GetSignature() []byte {
	return nct.Signature
}

var (
	ErrNegativeAmount  = fmt.Errorf("attempt to send negative funds")
	ErrIncorrectSigner = fmt.Errorf("attempt to sign from a different account")
)
