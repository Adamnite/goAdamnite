package vm

//file for general VM types and constants.
import (
	"encoding/binary"
	"math/big"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/params"
)

var LE = binary.LittleEndian

const (
	// DefaultPageSize is the linear memory page size.
	defaultPageSize = 65536
)

type (
	// CanTransferFunc is the signature of a transfer guard function
	CanTransferFunc func(*statedb.StateDB, common.Address, *big.Int) bool
	// TransferFunc is the signature of a transfer function
	TransferFunc func(*statedb.StateDB, common.Address, common.Address, *big.Int)
	// GetHashFunc returns the n'th block hash in the blockchain
	// and is used by the BLOCKHASH EVM op code.
	GetHashFunc func(uint64) common.Hash
)

type VirtualMachine interface {
	// step()
	run() error

	do()
	outputStack() string
}

type ControlBlock struct {
	code      []OperationCommon
	startAt   uint64
	elseAt    uint64
	endAt     uint64
	op        byte // Contains the value of the opcode that triggered this
	signature byte
	index     uint32
}
type Machine struct {
	VirtualMachine
	pointInCode       uint64
	contract          Contract
	vmCode            []OperationCommon
	vmStack           []uint64
	contractStorage   []uint64       //the storage of the smart contracts data.
	vmMemory          []byte         //i believe the agreed on stack size was
	locals            []uint64       //local vals that the VM code can call
	controlBlockStack []ControlBlock // Represents the labels indexes at which br, br_if can jump to
	config            VMConfig
	gas               uint64 // The allocated gas for the code execution
	callStack         []*Frame
	stopSignal        bool
	currentFrame      int
	BlockCtx          BlockContext
	txCtx             TxContext
	Statedb           *statedb.StateDB
	chainConfig       *params.ChainConfig
}

// BlockContext provides the EVM with auxiliary information. Once provided it shouldn't be modified.
type BlockContext struct {
	// CanTransfer returns whether the account contains
	// sufficient nite to transfer the value
	CanTransfer CanTransferFunc
	// Transfer transfers nite from one account to the other
	Transfer TransferFunc
	// GetHash returns the hash corresponding to n
	GetHash GetHashFunc

	// Block information
	Coinbase    common.Address
	GasLimit    uint64
	BlockNumber *big.Int
	Time        *big.Int
	Difficulty  *big.Int
	BaseFee     *big.Int
}

// TxContext provides the AdamniteVM with information about a transaction.
// All fields can change between transactions.
type TxContext struct {
	// Message information
	Origin   common.Address
	GasPrice *big.Int
}

type VMConfig struct {
	maxCallStackDepth        uint
	gasLimit                 uint64
	returnOnGasLimitExceeded bool
	debugStack               bool // should it output the stack every operation
	maxCodeSize              uint64
	codeGetter               GetCode
	codeBytesGetter          func(uri string, hash string) ([]byte, error)
	uri                      string
}

type Frame struct {
	Code         []OperationCommon
	Regs         []int64
	Locals       []uint64
	Ip           uint64
	ReturnReg    int
	Continuation int64
	CtrlStack    []ControlBlock
}

// Contract represents an adm contract in the state database. It contains
// the contract methods, calling arguments.
type Contract struct {
	Address       common.Address //the Address of the contract
	Value         *big.Int
	CallerAddress common.Address
	Code          []CodeStored
	Storage       []uint64
	Input         []byte // The bytes from `input` field of the transaction
	Gas           uint64
}
