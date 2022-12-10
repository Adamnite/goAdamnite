package vm

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
)

const (
	// DefaultPageSize is the linear memory page size.
	defaultPageSize = 65536
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
	module            Module // The module that will be executed inside the VM
	vmCode            []OperationCommon
	vmStack           []uint64
	contractStorage   []uint64       //the storage of the smart contracts data.
	vmMemory          []byte         //i believe the agreed on stack size was
	locals            []uint64       //local vals that the VM code can call
	controlBlockStack []ControlBlock // Represents the labels indexes at which br, br_if can jump to
	chainHandler      ChainDataHandler
	config            VMConfig
	gas               uint64 // The allocated gas for the code execution
	callStack         []*Frame
	stopSignal        bool
	currentFrame      int
}

type VMConfig struct {
	maxCallStackDepth        uint
	gasLimit                 uint64
	returnOnGasLimitExceeded bool
	debugStack               bool // should it output the stack every operation
	maxCodeSize              uint32
	codeGetter               GetCode
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

func getDefaultConfig() VMConfig {
	return VMConfig{
		maxCallStackDepth:        1024,
		gasLimit:                 30000, // 30000 ATE
		returnOnGasLimitExceeded: true,
		debugStack:               false,
		codeGetter:               defaultCodeGetter,
	}
}

type MemoryType interface {
	to_string() string
}
type Type_int64 struct {
	value int64
}

func (t Type_int64) to_string() string {
	return fmt.Sprint(t.value)
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

func newVirtualMachine(wasmBytes []byte, storage []uint64, config *VMConfig, gas uint64) *Machine {
	machine := new(Machine)
	machine.pointInCode = 0
	machine.contractStorage = storage
	machine.module = *decode(wasmBytes)
	// These delimited lines are left for compatibility purpose only should be removed
	machine.vmCode, machine.controlBlockStack = parseBytes(machine.module.codeSection[0].body)
	machine.locals = make([]uint64, len(machine.module.codeSection[0].localTypes))
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

	initMemoryWithDataSection(&machine.module, machine)
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

// The caller of call2 has to pass in the function hash from the `inputData` field
// The function identifier will be the first 4 bytes of the data and the remaining
// will be considered as function parameters

type GetCode func(hash []byte) (FunctionType, []OperationCommon, []ControlBlock)

func defaultCodeGetter(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
	panic(fmt.Errorf("virtual machine does not have a code getter setup"))
}

func (m *Machine) call2(callBytes string) {

	// Structure: 0x[16 bytes func identifer][param1..][param2...][param3]
	// Note: The callbytes is following the wasm encoding scheme.

	bytes, err := hex.DecodeString(callBytes)

	if err != nil {
		panic("Unable to parse bytes for call2")
	}

	// functionIdx, _, err := DecodeInt32(reader(bytes[0:]))

	// if err != nil {
	// 	panic("Error parsing function identifier for call2")
	// }

	// if int(functionIdx) > len(m.module.functionSection) {
	// 	panic("Call2 - No function with such index exists")
	// }

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
	m.locals = params
	m.vmCode, m.controlBlockStack = funcCode, controlStack

	currentFrame := m.callStack[m.currentFrame]
	currentFrame.Locals = m.locals
	currentFrame.Code = m.vmCode
	currentFrame.CtrlStack = m.controlBlockStack

	m.run()
}

func (m *Machine) reset() {
	//resets the stack, locals, and point in code. Also resets back to main frame
	m.locals = []uint64{}
	m.pointInCode = 0
	m.callStack[0].Ip = 0
	m.currentFrame = 0
	m.callStack[0].Code = m.vmCode
	m.callStack[0].CtrlStack = m.controlBlockStack
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
