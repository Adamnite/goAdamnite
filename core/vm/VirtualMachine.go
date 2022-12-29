package vm

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/params"
)

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

var LE = binary.LittleEndian

type VirtualMachine interface {
	//functions that will be fully implemented later
	step()
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
	contract		  Contract
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
	blockCtx		  BlockContext
	txCtx			  TxContext
	statedb    		  *statedb.StateDB
	chainConfig 	  *params.ChainConfig
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


// TxContext provides the EVM with information about a transaction.
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
	uri						 string
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
	Address common.Address   //the Address of the contract
	Value 	*big.Int
	CallerAddress common.Address
	Code 	[]CodeStored
	Storage []uint64
	Input   []byte // The bytes from `input` field of the transaction
	Gas 	uint64
}

func NewContract(caller common.Address, value *big.Int, input []byte, gas uint64) *Contract {
	c := &Contract{CallerAddress: caller, Value: value, Input: input, Gas: gas}
	return c
}


func GetCodeBytes2(uri string, hash string) ([]byte, error) {
	code, err := getMethodCode(uri, hash)
	if err != nil {
		return nil, err
	}
	return code.CodeBytes, nil
}


func getDefaultConfig() VMConfig {
	return VMConfig{
		maxCallStackDepth:        1024,
		gasLimit:                 30000, // 30000 ATE
		returnOnGasLimitExceeded: true,
		debugStack:               false,
		codeGetter:               defaultCodeGetter,
		codeBytesGetter: 		  GetCodeBytes2,
		uri:					  "https//default.uri",
	}
}


func (m *Machine) step() {
	if m.pointInCode < uint64(len(m.vmCode)) {
		op := m.vmCode[m.pointInCode]
		op.doOp(m)
		if m.config.debugStack {
			println(m.outputStack())
		}
	}
}

func (m *Machine) run() error {

	for m.currentFrame >= 0 {
		currentFrame := m.callStack[m.currentFrame]

		if currentFrame.Continuation != -1 {
			m.pointInCode = uint64(currentFrame.Continuation)
		}

		currentFrame.Ip = m.pointInCode
		m.vmCode = currentFrame.Code
		m.locals = currentFrame.Locals

		for uint64(currentFrame.Ip) < uint64(len(currentFrame.Code)) {
			oldFrameNum := m.currentFrame
			op := currentFrame.Code[currentFrame.Ip]
			err := op.doOp(m)

			if m.stopSignal {
				m.stopSignal = false
			}
			if m.config.debugStack {
				println(m.outputStack())
			}

			if err != nil {
				return err
			}
			currentFrame.Ip++

			// Activate the new frame
			if m.currentFrame > oldFrameNum {
				if m.currentFrame > int(m.config.maxCallStackDepth) {
					return ErrDepth
				}
				currentFrame = m.callStack[m.currentFrame]
			}
		}
		m.currentFrame--
	}
	return nil
}

func (m *Machine) outputStack() string {
	ans := ""
	for i, v := range m.vmStack {
		ans += strconv.FormatInt(int64(i), 16) + " ::: " + strconv.FormatUint(v, 16) + "\n"
		// println(fmt.Sprint(v))

	}
	return ans
}

func (m *Machine) outputMemory() string {
	ans := ""
	for _, v := range m.vmMemory {

		ans += strconv.FormatUint(uint64(v), 16) + " "
	}
	return ans
}

func initMemoryWithDataSection(module *Module, vm *Machine) {

	dataSegmentSize := uint32(len(module.dataSection))

	if module.dataCountSection != nil {
		dataSegmentSize = *module.dataCountSection
	}

	for i := uint32(0); i < uint32(dataSegmentSize); i++ {
		offset := module.dataSection[i].offsetExpression.data[0]
		size := len(module.dataSection[i].init)

		p := 0
		for j := uint32(offset); j < uint32(offset)+uint32(size); j++ {
			vm.vmMemory[j] = module.dataSection[i].init[p]
			p++
		}
	}
}

func initVMState(machine *Machine) {
	// Push the main frame
	machine.currentFrame = 0
	mainFrame := new(Frame)
	mainFrame.Ip = 0
	mainFrame.Continuation = -1
	mainFrame.Code = machine.vmCode
	mainFrame.CtrlStack = machine.controlBlockStack
	mainFrame.Locals = machine.locals
	machine.callStack = append(machine.callStack, mainFrame)

	capacity := 20 * defaultPageSize
	machine.vmMemory = make([]byte, capacity) // Initialize empty memory. (make creates array of 0)
	machine.locals = make([]uint64, 2)

	// Initialize memory with things inside the data section
	// initMemoryWithDataSection(&machine.module, machine)
}

func newVM(statedb *statedb.StateDB, bc BlockContext, txc TxContext, config* VMConfig, chainConfig* params.ChainConfig) *Machine {
	machine := new(Machine)
	machine.statedb = statedb
	machine.blockCtx = bc
	machine.txCtx = txc
	machine.chainConfig = chainConfig

	if config != nil {
		machine.config = *config
	} else {
		machine.config = getDefaultConfig()
	}

	return machine
}


func setCallCode(m* Machine, funcBodyBytes []byte, gas uint64) {
	m.vmCode, m.controlBlockStack = parseBytes(funcBodyBytes)
	m.gas = gas
}

func setCodeAndInit(m* Machine, funcBodyBytes []byte, gas uint64) {
	m.vmCode, m.controlBlockStack = parseBytes(funcBodyBytes)
	m.gas = gas
	initVMState(m)
}

// This constructor is let for compatibility only and should be updated/removed
func newVirtualMachine(wasmBytes []byte, storage []uint64, config *VMConfig, gas uint64) *Machine {
	machine := new(Machine)
	machine.pointInCode = 0
	machine.contractStorage = storage
	// machine.module = *decode(wasmBytes)
	// // These delimited lines are left for compatibility purpose only should be removed
	// machine.vmCode, machine.controlBlockStack = parseBytes(machine.module.codeSection[0].body)
	// machine.locals = make([]uint64, len(machine.module.codeSection[0].localTypes))
	//
	machine.gas = gas

	if config != nil {
		machine.config = *config
	} else {
		machine.config = getDefaultConfig()
	}

	// Push the main frame
	machine.currentFrame = 0

	mainFrame := new(Frame)
	mainFrame.Ip = 0
	mainFrame.Continuation = -1
	mainFrame.Code = machine.vmCode
	mainFrame.CtrlStack = machine.controlBlockStack
	mainFrame.Locals = machine.locals
	machine.callStack = append(machine.callStack, mainFrame)

	capacity := 20 * defaultPageSize
	machine.vmMemory = make([]byte, capacity) // Initialize empty memory. (make creates array of 0)

	// initMemoryWithDataSection(&machine.module, machine)
	// Initialize memory with things inside the data section
	return machine
}

func (m *Machine) popFromStack() uint64 {
	var ans uint64

	if m.config.debugStack {
		println("popping from stack")
	}
	ans, m.vmStack = m.vmStack[len(m.vmStack)-1], m.vmStack[:len(m.vmStack)-1]
	return ans
}
func (m *Machine) pushToStack(n interface{}) {

	switch v := n.(type) {
	case uint64:
		m.vmStack = append(m.vmStack, v)
	case uint32:
		m.vmStack = append(m.vmStack, uint64(v))
	case float64:
		m.vmStack = append(m.vmStack, math.Float64bits(v))
	case float32:
		m.vmStack = append(m.vmStack, uint64(math.Float32bits(v)))
	case int:
		m.vmStack = append(m.vmStack, uint64(v))
	default:
		fmt.Println(fmt.Errorf("unable to push type %T to stack, has value %v", n, v))
	}
	if m.config.debugStack {
		fmt.Printf("pushing to stack: %v\n", n)
	}

}

// useGas attempts the use of gas and subtracts it and returns true on success
func (m *Machine) useGas(gas uint64) bool {
	if m.gas < gas {
		return false
	}
	m.gas -= gas
	return true
}


type GetCode func(hash []byte) (FunctionType, []OperationCommon, []ControlBlock)

func defaultCodeGetter(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
	panic(fmt.Errorf("virtual machine does not have a code getter setup"))
}

// Called when invoking specific function inside the contract
func (m *Machine) call2(callBytes string, gas uint64) error {

	// Structure: 0x[16 bytes func identifer][param1..][param2...][param3]
	// Note: The callbytes is following the wasm encoding scheme.

	bytes, err := hex.DecodeString(callBytes)

	if err != nil {
		panic("Unable to parse bytes for call2")
	}

	funcIdentifier := bytes[:16]

	funcTypes, funcCode, controlStack := m.config.codeGetter(funcIdentifier)

	var params []uint64

	for i := uint64(len(funcIdentifier)); i < uint64(len(bytes)); i++ {

		valTypeByte := bytes[i]

		switch valTypeByte {
		case Op_i32_const:
			paramValue, count, err := DecodeInt32(reader(bytes[i+1:]))

			if err != nil {
				panic("Call2 - Error parsing function params i32")
			}

			params = append(params, uint64(paramValue))
			i += count

		case Op_i64_const:
			paramValue, count, err := DecodeInt64(reader(bytes[i+1:]))
			params = append(params, uint64(paramValue))

			if err != nil {
				panic("Call2 - Error parsing function params i64")
			}
			i += count

		case Op_f32_const:
			num := LE.Uint32(bytes[i+1 : 4])
			math.Float32frombits(num)
			params = append(params, uint64(num))
			i += 5

		case Op_f64_const:
			num := LE.Uint64(bytes[i+1:])
			math.Float64frombits(num)
			params = append(params, num)
			i += 9
		default:
			print("Parsed valtype %v", valTypeByte)
			panic("No such type known")
		}
	}

	expectedParamCount := len(funcTypes.params)
	incomingParamCount := len(params)

	if expectedParamCount != incomingParamCount {
		panic("Call2 - Param counts mismatch")
	}

	// Maybe Check the types of each params if they matches signature?
	callbytes, _ := hex.DecodeString(callBytes)
	setCodeAndInit(m, callbytes, gas)

	m.locals = params
	m.vmCode, m.controlBlockStack = funcCode, controlStack

	currentFrame := m.callStack[m.currentFrame]
	currentFrame.Locals = m.locals
	currentFrame.Code = m.vmCode
	currentFrame.CtrlStack = m.controlBlockStack
	return m.run()
}


// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (m* Machine) Call(caller common.Address, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	if m.currentFrame > int(m.config.maxCallStackDepth) {
		return nil, gas, ErrDepth
	}

	if value.Sign() != 0 && !m.blockCtx.CanTransfer(m.statedb, caller, value) {
		return nil, gas, ErrInsufficientBalance
	}

	if !m.statedb.Exist(addr) {
		m.statedb.CreateAccount(addr)
	}

	snapshot := m.statedb.Snapshot()
	m.blockCtx.Transfer(m.statedb, caller, addr, value)

	// Retrieve the method code and execute it 
	if len(input) == 0 {
		return nil, gas, nil
	}

	contract := NewContract(caller, value, input, gas)
	m.contract = *contract

	err = m.call2(string(input), gas)

	if err != nil {
		m.statedb.RevertToSnapshot(snapshot)
		if err != ErrExecutionReverted {
			m.gas = 0
		}

		return nil, m.gas, err
	}
	return nil, m.gas, err
}


func getModuleLen(module *Module) uint64 {
	le := uint64(0)
	for i := uint64(0); i < uint64(len(module.functionSection)); i++ {
		le += uint64(len(module.codeSection[i].body))
	}
	return le
}

func (m *Machine) create(caller common.Address, codeBytes []byte, gas uint64, value *big.Int, address common.Address) ([]byte, common.Address, uint64, error) {

	if m.currentFrame > int(m.config.maxCallStackDepth) {
		return nil, common.Address{}, gas, ErrDepth
	}

	if !m.blockCtx.CanTransfer(m.statedb, caller, value) {
		return nil, common.Address{}, gas, ErrInsufficientBalance
	}

	nonce := m.statedb.GetNonce(caller)

	if nonce+1 < nonce {
		return nil, common.Address{}, gas, ErrNonceUintOverflow
	}

	m.statedb.SetNonce(caller, nonce+1)
 
	// Ensure there's no existing contract already at the designated address
	contractData, err := getContractData(m.config.uri, address.String())

	if err != nil {
		return nil, address, gas, err
	}

	if m.statedb.GetNonce(address) != 0 || contractData.Address != "" {
		return nil, common.Address{}, 0, ErrContractAddressCollision
	}

	// Create a new account on the state
	snapshot := m.statedb.Snapshot()
	m.statedb.CreateAccount(address)

	m.blockCtx.Transfer(m.statedb, caller, address, value)

	// Initialise a new contract and set the code that is to be used by the EVM.
	contract := NewContract(caller, value, codeBytes, gas)
	m.contract = *contract

	module := *decode(codeBytes)
	modLen := getModuleLen(&module)

	// Check whether the max code size has been exceeded
	if err == nil && modLen > m.config.maxCodeSize {
		m.statedb.RevertToSnapshot(snapshot)
		return nil, address, 0, ErrMaxCodeSizeExceeded
	}

	// if the contract creation ran successfully and no errors were returned
	// calculate the gas required to store the code.

	// @TODO update this with right creation price
	createModuleGas := modLen * 2000
	if createModuleGas > m.gas {
		m.statedb.RevertToSnapshot(snapshot)
		return nil, address, m.gas, ErrCodeStoreOutOfGas
	}

	// Upload the module here
	_, _, err = uploadModuleFunctions(m.config.uri, module)

	if err != nil {
		m.statedb.RevertToSnapshot(snapshot)
		// Somehow revert the uploading here?
	}

	m.gas -= createModuleGas
	return nil, address, contract.Gas, err
}

// Create creates a new contract using code as deployment code.
func (m *Machine) Create(caller common.Address, code []byte, gas uint64, value *big.Int) (ret []byte, contractAddr common.Address, leftOverGas uint64, err error) {
	addrBytes := caller.Bytes()
	nonce := m.statedb.GetNonce(caller)
	addrBytes = append(addrBytes, byte(nonce))

	contractAddr = crypto.PubkeyByteToAddress(addrBytes)
	return m.create(caller, code, gas, value, contractAddr)
}

func (m *Machine) reset() {
	//resets the stack, locals, and point in code. Also resets back to main frame
	m.locals = []uint64{}
	m.pointInCode = 0
	m.callStack[0].Ip = 0
	m.currentFrame = 0
	m.callStack[0].Code = m.vmCode
	m.callStack[0].CtrlStack = m.controlBlockStack
	m.callStack[0].Locals = m.locals
	m.callStack = []*Frame{m.callStack[0]}
}

func (m *Machine) addLocal(n interface{}) {
	switch v := n.(type) {
	case uint64:
		m.locals = append(m.locals, v)
	case uint32:
		m.locals = append(m.locals, uint64(v))
	case float64:
		m.locals = append(m.locals, math.Float64bits(v))
	case float32:
		m.locals = append(m.locals, uint64(math.Float32bits(v)))
	case []uint64:
		for i := 0; i < len(v); i++ {
			m.addLocal(v[i])
		}
	case []uint32:
		for i := 0; i < len(v); i++ {
			m.addLocal(v[i])
		}
	case []float64:
		for i := 0; i < len(v); i++ {
			m.addLocal(v[i])
		}
	case []float32:
		for i := 0; i < len(v); i++ {
			m.addLocal(v[i])
		}

	default:
		fmt.Println(fmt.Errorf("unable to push type %T to locals, has value %v", n, v))
	}
	if m.config.debugStack {
		fmt.Printf("pushed to locals: %v\n", n)
	}
}
