package VM

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/params"
)

func newContract(caller common.Address, value *big.Int, input []byte, gas uint64) *Contract {
	c := &Contract{CallerAddress: caller, Value: value, Input: input, Gas: gas}
	return c
}

func GetCodeBytes2(uri string, hash string) ([]byte, error) {
	code, err := GetMethodCode(uri, hash)
	if err != nil {
		return nil, err
	}
	return code.CodeBytes, nil
}

func GetDefaultConfig() VMConfig {
	return VMConfig{
		maxCallStackDepth:        1024,
		gasLimit:                 30000, // 30000 ATE
		returnOnGasLimitExceeded: true,
		debugStack:               false,
		CodeGetter:               defaultCodeGetter,
		CodeBytesGetter:          GetCodeBytes2,
		Uri:                      "https//default.uri",
	}
}

func (m *Machine) step() {
	//DANGER!!! Untested! No longer same as an individual Run step!
	if m.pointInCode < uint64(len(m.vmCode)) {
		op := m.vmCode[m.pointInCode]
		op.doOp(m)
		m.debugOutputStack()
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
			m.debugOutputStack()

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

func (m *Machine) debugOutputStack() {
	//removing repeated code with a function that outputs the stack if debug is set to True
	if m.config.debugStack {
		fmt.Println(m.OutputStack())
	}
}

func (m *Machine) OutputStack() string {
	ans := ""
	for i, v := range m.vmStack {
		ans += strconv.FormatInt(int64(i), 16) + " ::: " + strconv.FormatUint(v, 16) + "\n"
		// println(fmt.Sprint(v))

	}
	return ans
}

func (m *Machine) OutputMemory() string {
	ans := ""
	for _, v := range m.vmMemory {

		ans += strconv.FormatUint(uint64(v), 16) + " "
	}
	return ans
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
func CanTransfer(db *statedb.StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db *statedb.StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}

func NewBlockContext(coinbase common.Address, ateLimit uint64, blockNumber *big.Int, time *big.Int, diff *big.Int, fee *big.Int) BlockContext {
	bc := BlockContext{}
	bc.Coinbase = coinbase
	bc.GasLimit = ateLimit
	bc.BlockNumber = blockNumber
	bc.Time = time
	bc.Difficulty = diff
	bc.BaseFee = fee
	bc.CanTransfer = CanTransfer
	bc.Transfer = Transfer

	return bc
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

func NewVM(statedb *statedb.StateDB, bc BlockContext, txc TxContext, config *VMConfig, chainConfig *params.ChainConfig) *Machine {
	machine := new(Machine)
	machine.Statedb = statedb
	machine.BlockCtx = bc
	machine.txCtx = txc
	machine.chainConfig = chainConfig

	if config != nil {
		machine.config = *config
	} else {
		machine.config = GetDefaultConfig()
	}
	machine.gas = config.gasLimit //TODO: check this out

	return machine
}

func SetCallCode(m *Machine, funcBodyBytes []byte, gas uint64) {
	m.vmCode, m.controlBlockStack = parseBytes(funcBodyBytes)
	m.gas = gas
}

func SetCodeAndInit(m *Machine, funcBodyBytes []byte, gas uint64) {
	m.vmCode, m.controlBlockStack = parseBytes(funcBodyBytes)
	m.gas = gas
	initVMState(m)
}

// This constructor is let for compatibility only and should be updated/removed
func NewVirtualMachine(wasmBytes []byte, storage []uint64, config *VMConfig, gas uint64) *Machine {
	machine := new(Machine)
	machine.pointInCode = 0
	machine.contractStorage = storage
	machine.gas = gas

	if config != nil {
		machine.config = *config
	} else {
		machine.config = GetDefaultConfig()
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
	machine.storageChanges = map[uint32]uint64{}

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

func (m *Machine) DumpStack() {
	fmt.Printf("Stack Output: %v\n", m.vmStack)
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

// useAte attempts the use of ate and subtracts it and returns true on success
func (m *Machine) useAte(gas uint64) bool {
	if m.gas < gas {
		return false
	}
	m.gas -= gas
	return true
}

func defaultCodeGetter(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
	panic(fmt.Errorf("virtual machine does not have a code getter setup"))
}

// Called when invoking specific function inside the contract
func (m *Machine) Call2(callBytes interface{}, gas uint64) error {
	// Structure: 0x[16 bytes func identifier][param1..][param2...][param3]
	// Note: The callbytes is following the wasm encoding scheme. can be passed as string or byte array
	var bytes []byte
	switch v := callBytes.(type) {
	case string:
		var err error
		bytes, err = hex.DecodeString(v)
		if err != nil {
			return fmt.Errorf("unable to parse bytes from string for call2")
		}
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("unable to parse bytes from %v for call2", v)
	}

	funcIdentifier := bytes[:16]
	funcTypes, funcCode, controlStack := m.config.CodeGetter(funcIdentifier)
	var params []uint64
	//get the params from the bytes passed.
	for i := uint64(len(funcIdentifier)); i < uint64(len(bytes)); i++ {

		valTypeByte := bytes[i]

		switch valTypeByte {
		case Op_i32:
			paramValue, count, err := DecodeInt32(reader(bytes[i+1:]))

			if err != nil {
				return fmt.Errorf("call2 - Error parsing function params i32")
			}

			params = append(params, uint64(paramValue))
			i += count

		case Op_i64:
			paramValue, count, err := DecodeInt64(reader(bytes[i+1:]))
			params = append(params, uint64(paramValue))

			if err != nil {
				return fmt.Errorf("call2 - Error parsing function params i64")
			}
			i += count

		case Op_f32:
			num := LE.Uint32(bytes[i+1 : 4])
			math.Float32frombits(num)
			params = append(params, uint64(num))
			i += 4

		case Op_f64:
			num := LE.Uint64(bytes[i+1:])
			fmt.Println("num version of f64 is: ", num)
			fmt.Println(math.Float64frombits(num))
			params = append(params, num)
			i += 8
		default:
			println("Parsed valtype %v", valTypeByte)
			println("at index ", i)
			return fmt.Errorf("parsed valtype %v, no such known type", valTypeByte)
		}
	}

	expectedParamCount := len(funcTypes.params)
	incomingParamCount := len(params)

	if expectedParamCount != incomingParamCount {
		fmt.Printf("Expecting: %v Got %v ", expectedParamCount, incomingParamCount)
		panic("Call2 - Param counts mismatch")
	}

	// Maybe Check the types of each params if they matches signature?
	//shouldn't this be using the function body???

	// setCodeAndInit(m, bytes, gas)
	m.gas = gas
	initVMState(m)

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
func (m *Machine) Call(caller common.Address, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	if m.currentFrame > int(m.config.maxCallStackDepth) {
		return nil, gas, ErrDepth
	}

	if value.Sign() != 0 && !m.BlockCtx.CanTransfer(m.Statedb, caller, value) {
		return nil, gas, ErrInsufficientBalance
	}

	if !m.Statedb.Exist(addr) {
		m.Statedb.CreateAccount(addr)
	}

	snapshot := m.Statedb.Snapshot()
	m.BlockCtx.Transfer(m.Statedb, caller, addr, value)

	// Retrieve the method code and execute it
	if len(input) == 0 {
		return nil, gas, nil
	}

	contract := newContract(caller, value, input, gas)
	m.contract = *contract

	err = m.Call2(input, gas)

	if err != nil {
		m.Statedb.RevertToSnapshot(snapshot)
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

func (m *Machine) create(caller common.Address, codeBytes []byte, gas uint64, value *big.Int, address common.Address) (common.Address, uint64, error) {
	//creates a new contract from the bytes passed (whole module) and uploads those functions, then has a locally created contract (only in this VM)
	if m.currentFrame > int(m.config.maxCallStackDepth) {
		return common.Address{}, gas, ErrDepth
	}

	if !m.BlockCtx.CanTransfer(m.Statedb, caller, value) {
		return common.Address{}, gas, ErrInsufficientBalance
	}

	nonce := m.Statedb.GetNonce(caller)

	if nonce+1 < nonce {
		return common.Address{}, gas, ErrNonceUintOverflow
	}

	m.Statedb.SetNonce(caller, nonce+1)

	// Ensure there's no existing contract already at the designated address
	_, err := GetContractData(m.config.Uri, address.Hex())

	if err == nil || err != ERR_CONTRACT_NOT_STORED {
		//either theres a strange error, or the contract already exists.
		return address, gas, err
	}

	if m.Statedb.GetNonce(address) != 0 || address.String() == "" {
		return common.Address{}, 0, ErrContractAddressCollision
	}

	// Create a new account on the state
	snapshot := m.Statedb.Snapshot()
	m.Statedb.CreateAccount(address)

	m.BlockCtx.Transfer(m.Statedb, caller, address, value)

	// Initialize a new contract and set the code that is to be used by the ADVM.
	contract := newContract(caller, value, codeBytes, gas)
	contract.Address = address
	m.contract = *contract

	module := *decode(codeBytes)
	modLen := getModuleLen(&module)

	// Check whether the max code size has been exceeded
	if err == nil && modLen > m.config.maxCodeSize {
		m.Statedb.RevertToSnapshot(snapshot)
		return address, 0, ErrMaxCodeSizeExceeded
	}

	// if the contract creation ran successfully and no errors were returned
	// calculate the gas required to store the code.

	// @TODO update this with right creation price
	//TODO: update this to charge only for new methods being uploaded.
	createModuleGas := modLen * 2000
	if createModuleGas > m.gas {
		m.Statedb.RevertToSnapshot(snapshot)
		return address, m.gas, ErrCodeStoreOutOfGas
	}

	// Upload the module here
	codeStored, hashesOfMethods, err := UploadModuleFunctions(m.config.Uri, module)

	if err != nil {
		m.Statedb.RevertToSnapshot(snapshot)
		// Somehow revert the uploading here?
	}
	m.contract.Code = codeStored
	m.contract.CodeHashes = make([]string, len(hashesOfMethods))
	for i, hashBytes := range hashesOfMethods {
		m.contract.CodeHashes[i] = hex.EncodeToString(hashBytes)
	}

	m.gas -= createModuleGas
	return address, contract.Gas, err
}

// Create creates a new contract using code as deployment code.
func (m *Machine) Create(caller common.Address, code []byte, gas uint64, value *big.Int) (contractAddr common.Address, leftOverGas uint64, err error) {
	addrBytes := caller.Bytes()
	nonce := m.Statedb.GetNonce(caller)
	addrBytes = append(addrBytes, byte(nonce))

	contractKey, err := crypto.GenerateKey()

	contractAddr = crypto.PubkeyToAddress(contractKey.PublicKey)

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

func (m *Machine) AddLocal(n interface{}) {
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
			m.AddLocal(v[i])
		}
	case []uint32:
		for i := 0; i < len(v); i++ {
			m.AddLocal(v[i])
		}
	case []float64:
		for i := 0; i < len(v); i++ {
			m.AddLocal(v[i])
		}
	case []float32:
		for i := 0; i < len(v); i++ {
			m.AddLocal(v[i])
		}

	default:
		fmt.Println(fmt.Errorf("unable to push type %T to locals, has value %v", n, v))
	}
	if m.config.debugStack {
		fmt.Printf("pushed to locals: %v\n", n)
	}
}

func (m *Machine) getContractHash() common.Hash {
	return m.contract.Hash()
}

func (m *Machine) UploadContract(APIEndpoint string) error {
	//takes the contract (or just its changes) and uploads that to the DB at the APIEndpoint link
	return UploadContract(APIEndpoint, m.contract)
}

// sort the changes from the method being ran, allowing this to be passed along the chain and to the OffChainDB efficiently
func (m *Machine) GetChanges() RuntimeChanges {
	// we want to return an array of the changes, with as few starts as possible,
	changes := RuntimeChanges{
		Caller:            m.contract.CallerAddress,
		CallTime:          time.Now().UTC(), //this should not be used! this should be changed to the time at the start of the call.
		ContractCalled:    m.contract.Address,
		ParametersPassed:  m.locals,
		ChangeStartPoints: []uint64{},
		Changed:           [][]byte{},
	}
	keys := make([]uint32, 0, len(m.storageChanges))
	for k := range m.storageChanges {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, point := range keys {
		final := m.storageChanges[point]
		finalBytes := LE.AppendUint64([]byte{}, final)
		bytePoint := point * 8 //point stores relative to uint64s, this is using bytes, so we need to change that
		if len(changes.ChangeStartPoints) > 0 &&
			int(changes.ChangeStartPoints[len(changes.ChangeStartPoints)-1])+
				len(changes.Changed[len(changes.Changed)-1]) ==
				int(bytePoint) {

			changes.Changed[len(changes.Changed)-1] = append(changes.Changed[len(changes.Changed)-1], finalBytes...)

		} else {
			changes.ChangeStartPoints = append(changes.ChangeStartPoints, uint64(bytePoint))
			changes.Changed = append(changes.Changed, finalBytes)
		}
	}
	if m.config.debugStack {
		changes.OutputChanges()
	}
	return changes
}
