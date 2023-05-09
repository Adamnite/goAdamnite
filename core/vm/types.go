package VM

//file for general VM types and constants.
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"time"

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
	Run() error

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
	vmStack           []uint64          //the stack the VM uses
	contractStorage   []uint64          //the storage of the smart contracts data.
	storageChanges    map[uint32]uint64 //point to new value
	vmMemory          []byte            //i believe the agreed on stack size was
	locals            []uint64          //local vals that the VM code can call
	controlBlockStack []ControlBlock    // Represents the labels indexes at which br, br_if can jump to
	config            VMConfig
	gas               uint64 // The allocated gas for the code execution
	callStack         []*Frame
	stopSignal        bool
	currentFrame      int
	BlockCtx          BlockContext
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

type GetCode func(hash []byte) (FunctionType, []OperationCommon, []ControlBlock)

type VMConfig struct {
	maxCallStackDepth        uint
	gasLimit                 uint64
	returnOnGasLimitExceeded bool
	debugStack               bool // should it output the stack every operation
	maxCodeSize              uint64
	CodeGetter               GetCode
	CodeBytesGetter          func(uri string, hash string) ([]byte, error)
	Uri                      string
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
	CodeHashes    []string //the hash of the code,the code is only actually in Contract.Code once its called
	Storage       []uint64
	Input         []byte // The bytes from `input` field of the transaction
	Gas           uint64
}

type RuntimeChanges struct {
	Caller            common.Address //who called this
	CallTime          time.Time      //when was it called
	ContractCalled    common.Address //what was called
	ParametersPassed  []byte         //what was passed on the call
	GasLimit          uint64         //did they set a gas limit
	ChangeStartPoints []uint64       //data from the results
	Changed           [][]byte       //^
	ErrorsEncountered error          //if anything went wrong at runtime
}

func (rtc RuntimeChanges) OutputChanges() {
	for i, x := range rtc.ChangeStartPoints {
		dataString := "\n"
		for foo := 0; foo < len(rtc.Changed[i]); foo += 8 {
			dataString = fmt.Sprintf(
				"%v\t (%3v) \t0x % X  \n",
				dataString, LE.Uint64(rtc.Changed[i][foo:foo+8]),
				rtc.Changed[i][foo:foo+8],
			)
		}
		log.Printf("change at real index %v, byte index: %v, changed to: %v", x/8, x, dataString)

	}
	log.Println()
}

// returns a copy without any changes. Ideal for running again to check consistency
func (rtc RuntimeChanges) CleanCopy() *RuntimeChanges {
	ans := RuntimeChanges{
		Caller:            rtc.Caller,
		ContractCalled:    rtc.ContractCalled,
		GasLimit:          rtc.GasLimit,
		ChangeStartPoints: []uint64{},
		Changed:           [][]byte{},
		ErrorsEncountered: nil,
	}
	return &ans
}

func (a RuntimeChanges) Equal(b RuntimeChanges) bool {
	if bytes.Equal(a.Caller.Bytes(), b.Caller.Bytes()) &&
		bytes.Equal(a.ContractCalled[:], b.ContractCalled[:]) &&
		a.GasLimit == b.GasLimit && a.ErrorsEncountered == b.ErrorsEncountered &&
		len(a.ChangeStartPoints) == len(b.ChangeStartPoints) &&
		len(a.Changed) == len(b.Changed) {

		//check the change start points
		for i, aStart := range a.ChangeStartPoints {
			if b.ChangeStartPoints[i] != aStart {
				return false
			}
		}
		//check the changed values
		for i, aChanged := range a.Changed {
			if !bytes.Equal(aChanged, b.Changed[i]) {
				return false
			}
		}

		return true
	}
	return false
}
