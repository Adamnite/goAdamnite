package VM

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"
)

func NewVirtualMachineWithSpoofedConnection(contract *common.Address, dbInterface DBInterfaceItem) (*Machine, error) {
	vm := NewVirtualMachine([]byte{}, []uint64{}, 1000)
	vm.config.Getter = dbInterface
	if contract == nil {
		return vm, ErrContractNotStored
	}
	//get the contract from the DB
	con, err := dbInterface.GetContract(contract.Hex())
	if err != nil {
		return vm, err
	}
	vm.contract = *con

	return vm, nil
}
func NewVirtualMachineWithContract(apiEndpoint string, contract *common.Address) (*Machine, error) {
	vm := NewVirtualMachine([]byte{}, []uint64{}, 1000)

	if contract == nil {
		return vm, ErrContractNotStored
	}
	if err := vm.ResetToContract(apiEndpoint, *contract); err != nil {
		return nil, err
	}

	return vm, nil
}
func (vm *Machine) ResetToContract(apiEndpoint string, contract common.Address) error {
	vm.Reset()
	con, err := GetContractData(apiEndpoint, contract.Hex())
	if err != nil {
		return err
	}
	vm.contract = *con
	return nil
}
func (vm *Machine) CallWith(apiEndpoint string, rt *utils.RuntimeChanges) (*utils.RuntimeChanges, error) {
	if err := vm.ResetToContract(apiEndpoint, rt.ContractCalled); err != nil {
		return nil, err
	}
	// spoofer := NewDBSpoofer() //this is either smart, or *very* stupid, i cant honestly tell
	//TODO: either way, cleanup here
	// vm.config.Getter. = func(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
	// 	stored, err := GetMethodCode(apiEndpoint, hex.EncodeToString(hash))
	// 	if err != nil {
	// 		panic(err) //TODO: handle this better.
	// 	}
	// 	spoofer.AddSpoofedCode(hex.EncodeToString(hash), *stored)

	// 	return spoofer.GetCode(hash)
	// }
	return vm.CallOnContractWith(rt)
}
func (vm *Machine) CallOnContractWith(rt *utils.RuntimeChanges) (*utils.RuntimeChanges, error) {
	err := vm.Call2(rt.ParametersPassed, rt.GasLimit)
	rt.ErrorsEncountered = err
	return vm.UpdateChanges(rt), err
}
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
		debugStack: false,
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
				if m.currentFrame > maxCallStackDepth {
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
func NewVirtualMachine(wasmBytes []byte, storage []uint64, gas uint64) *Machine {
	machine := new(Machine)
	machine.pointInCode = 0
	machine.contractStorage = storage
	machine.gas = gas
	machine.config = GetDefaultConfig()

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
	// Note: The call bytes is following the wasm encoding scheme. can be passed as string or byte array
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
	funcTypes, funcCode, controlStack := m.config.Getter.GetCode(funcIdentifier)
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
			println("Parsed val type %v", valTypeByte)
			println("at index ", i)
			return fmt.Errorf("parsed val type %v, no such known type", valTypeByte)
		}
	}

	expectedParamCount := len(funcTypes.params)
	incomingParamCount := len(params)

	if expectedParamCount != incomingParamCount {
		fmt.Printf("Expecting: %v Got %v ", expectedParamCount, incomingParamCount)
		return fmt.Errorf("Call2 - Param counts mismatch")
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
	if m.currentFrame > maxCallStackDepth {
		return nil, gas, ErrDepth
	}

	if !m.Statedb.Exist(addr) {
		m.Statedb.CreateAccount(addr)
	}

	snapshot := m.Statedb.Snapshot()

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

func (m *Machine) Reset() {
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

func (m *Machine) GetContractHash() common.Hash {
	return m.contract.Hash()
}

func (m *Machine) UploadContract(APIEndpoint string) error {
	//takes the contract (or just its changes) and uploads that to the DB at the APIEndpoint link
	return UploadContract(APIEndpoint, m.contract)
}

// sort the changes from the method being ran, allowing this to be passed along the chain and to the OffChainDB efficiently
func (m *Machine) GetChanges() utils.RuntimeChanges {
	// we want to return an array of the changes, with as few starts as possible,
	changes := utils.RuntimeChanges{
		Caller:            m.contract.CallerAddress,
		CallTime:          time.Now().UTC(), //this should not be used! this should be changed to the time at the start of the call.
		ContractCalled:    m.contract.Address,
		ChangeStartPoints: []uint64{},
		Changed:           [][]byte{},
	}
	return *m.UpdateChanges(&changes)
}

func (m *Machine) UpdateChanges(changes *utils.RuntimeChanges) *utils.RuntimeChanges {
	changes.Changed = [][]byte{}
	changes.ChangeStartPoints = []uint64{}
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
		bytePoint := uint64(point) * 8 //point stores relative to uint64s, this is using bytes, so we need to change that
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
