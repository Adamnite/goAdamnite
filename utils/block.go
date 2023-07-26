package utils

//TODO: once this block is the only version, it (as well as transactions, and similar files) should be moved to RPC
import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

type BlockHeader struct {
	Timestamp time.Time // Timestamp at which the block was approved as Unix time stamp

	ParentBlockID         common.Hash      // Hash of the parent block
	Witness               crypto.PublicKey // Address of the witness who proposed the block
	WitnessMerkleRoot     common.Hash      // Merkle tree root in which witnesses for this block are stored
	TransactionMerkleRoot common.Hash      // Merkle tree root in which transactions for this block are stored
	StateMerkleRoot       common.Hash      // Merkle tree root in which states for this block are stored
	TransactionType       int8
	Number                *big.Int //the block number
	Round                 uint64
}

func NewBlockHeader(parentHash common.Hash, witness crypto.PublicKey, witnessRoot common.Hash, transactionRoot common.Hash, stateRoot common.Hash, transactionType int8, blockNumber *big.Int, roundNumber uint64) *BlockHeader {
	bh := BlockHeader{
		Timestamp:             time.Now().UTC(),
		ParentBlockID:         parentHash,
		Witness:               witness,
		WitnessMerkleRoot:     witnessRoot,
		TransactionMerkleRoot: transactionRoot,
		StateMerkleRoot:       stateRoot,
		TransactionType:       transactionType,
		Number:                blockNumber,
		Round:                 roundNumber,
	}
	return &bh
}
func NextBlockHeader(parent BlockHeader, witness crypto.PublicKey, witnessRoot common.Hash, transactionRoot common.Hash, stateRoot common.Hash, roundNumber uint64) *BlockHeader {
	bh := BlockHeader{
		Timestamp:             time.Now().UTC(),
		ParentBlockID:         parent.Hash(),
		Witness:               witness,
		WitnessMerkleRoot:     witnessRoot,
		TransactionMerkleRoot: transactionRoot,
		StateMerkleRoot:       stateRoot,
		TransactionType:       parent.TransactionType,
		Number:                big.NewInt(0).Add(parent.Number, big.NewInt(1)),
		Round:                 roundNumber,
	}
	return &bh
}

// Hash hashes block header
func (h *BlockHeader) Hash() common.Hash {
	//after being marshalled, the timezone will be set to values of nil, and not just nil. this breaks
	//a surprising amount of things
	val := binary.BigEndian.AppendUint64([]byte{}, uint64(h.Timestamp.UTC().UnixMilli()))
	val = append(val, h.ParentBlockID[:]...)
	val = append(val, h.Witness[:]...)
	val = append(val, h.WitnessMerkleRoot[:]...)
	val = append(val, h.StateMerkleRoot[:]...)
	val = append(val, h.TransactionMerkleRoot[:]...)
	val = append(val, h.Number.Bytes()...)
	val = binary.BigEndian.AppendUint64(val, h.Round)

	sha := sha256.New()
	sha.Write(val)
	return common.BytesToHash(sha.Sum(nil))
}

type BlockType interface {
	Hash() common.Hash
	Sign(accounts.Account) error
	VerifySignature() bool
	GetHeader() *BlockHeader
}

type Block struct {
	Header       *BlockHeader
	Transactions []*BaseTransaction
	Signature    []byte
}

// NewBlock creates and returns Block
func NewBlock(parentBlockID common.Hash, witness crypto.PublicKey, witnessRoot common.Hash, transactionRoot common.Hash, stateRoot common.Hash, number *big.Int, transactions []*BaseTransaction) *Block {
	header := &BlockHeader{
		Timestamp:             time.Now().UTC(),
		ParentBlockID:         parentBlockID,
		Witness:               witness,
		WitnessMerkleRoot:     witnessRoot,
		TransactionMerkleRoot: transactionRoot,
		StateMerkleRoot:       stateRoot,
		Number:                number,
	}
	header.TransactionType = Transaction_Basic //TODO: make this correct
	block := &Block{
		Header:       header,
		Transactions: transactions,
	}
	return block
}
func NextBlock(parent BlockHeader, witness crypto.PublicKey, witnessRoot common.Hash, transactionRoot common.Hash, stateRoot common.Hash, round uint64, transactions []*BaseTransaction) *Block {
	header := NextBlockHeader(parent, witness, witnessRoot, transactionRoot, stateRoot, round)
	block := &Block{
		Header:       header,
		Transactions: transactions,
	}
	return block
}

// Hash gets block's hash
func (b *Block) Hash() common.Hash {
	bytes := b.Header.Hash().Bytes()
	for _, t := range b.Transactions {
		bytes = append(bytes, (t).GetSignature()...)
		//since the signature normally isn't added to the transactions hash (how would you sign the hash of a signature of a hash of a....)
	}
	return common.BytesToHash(crypto.Sha512(bytes))
}

func (b *Block) Sign(signer accounts.Account) error {
	sig, err := signer.Sign(b)
	b.Signature = sig
	return err
}

// returns true if the signature is valid
func (b *Block) VerifySignature() bool {
	if b.Header == nil || b.Header.Witness == nil || len(b.Signature) < 64 {
		return false //might as well add safe guards
	}
	signer := accounts.AccountFromPubBytes(b.Header.Witness)
	return signer.Verify(b, b.Signature)
}
func (b *Block) GetHeader() *BlockHeader {
	return b.Header
}

type VMBlock struct {
	Header         *BlockHeader
	Transactions   []TransactionType
	BalanceChanges map[common.Address]*big.Int
	HashChanges    map[common.Address][2]common.Hash //contractAddress->[start, end]
	Signature      []byte

	//items that are only used in formation of the block
	lock sync.Mutex
}

func NewWorkingVMBlock(parentBlockID common.Hash, witness crypto.PublicKey, witnessRoot common.Hash, transactionRoot common.Hash, stateRoot common.Hash, number *big.Int, round uint64, transactions []TransactionType) *VMBlock {
	bh := NewBlockHeader(parentBlockID, witness, witnessRoot, transactionRoot, stateRoot, Transaction_VM_Call, number, round)
	vmb := VMBlock{
		Header:         bh,
		Transactions:   transactions,
		BalanceChanges: make(map[common.Address]*big.Int),
		HashChanges:    make(map[common.Address][2]common.Hash),
		lock:           sync.Mutex{},
	}
	return &vmb
}
func (vmb *VMBlock) Hash() common.Hash {
	data := vmb.Header.Hash().Bytes()
	for _, t := range vmb.Transactions {
		data = append(data, t.GetSignature()...)
	}
	for from, change := range vmb.BalanceChanges {
		data = append(data, from.Bytes()...)
		data = append(data, change.Bytes()...)
	}
	for contract, hashes := range vmb.HashChanges {
		data = append(data, contract.Bytes()...)
		data = append(data, hashes[0].Bytes()...)
		data = append(data, hashes[1].Bytes()...)
	}
	return common.BytesToHash(crypto.Sha512(data))
}

func (vmb *VMBlock) Sign(signer accounts.Account) error {
	sig, err := signer.Sign(vmb)
	vmb.Signature = sig
	return err
}
func (vmb *VMBlock) VerifySignature() bool {
	if vmb.Header == nil || vmb.Header.Witness == nil {
		return false //might as well add safe guards
	}
	signer := accounts.AccountFromPubBytes(vmb.Header.Witness)
	return signer.Verify(vmb, vmb.Signature)
}
func (vmb *VMBlock) GetHeader() *BlockHeader {
	return vmb.Header
}

// should be 0 on any valid blocks.
func (vmb *VMBlock) GetTransferSum() *big.Int {
	ans := big.NewInt(0)
	for _, amount := range vmb.BalanceChanges {
		ans = ans.Add(ans, amount)
	}
	return ans
}
func (vmb *VMBlock) ApplyTransfersTo(state *statedb.StateDB) error {
	if vmb.GetTransferSum().Cmp(big.NewInt(0)) != 0 {
		return fmt.Errorf("balances do not line up")
	}
	for address, total := range vmb.BalanceChanges {
		if state.GetBalance(address).Cmp(big.NewInt(0).Mul(big.NewInt(-1), total)) == -1 {
			//times -1 so this only maters net loss of funds.
			return ErrNegativeAmount //TODO: replace with a better error
		}
		state.AddBalance(address, total)
	}
	return nil
}
