package vm

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"math"
	"encoding/hex"
)

const (
	// DefaultPageSize is the linear memory page size.
	defaultPageSize = 65536
)

var LE = binary.LittleEndian

type VirtualMachine interface {
	//functions that will be fully implemented later
	step()
	run()

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
	contractStorage   []byte   //the storage of the smart contracts data.
	vmMemory          []byte   //i believe the agreed on stack size was
	locals            []uint64 //local vals that the VM code can call
	debugStack        bool     //should it output the stack every operation
	globals           []uint64
	controlBlockStack []ControlBlock // Represents the labels indexes at which br, br_if can jump to

	config VMConfig
	gas    uint64 // The allocated gas for the code execution
}

type VMConfig struct {
	maxCallStackDepth        uint
	gasLimit                 uint64
	returnOnGasLimitExceeded bool
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
		if m.debugStack {
			println(m.outputStack())
		}
	}
}

func (m *Machine) run() {
	for m.pointInCode < uint64(len(m.vmCode)) {
		op := m.vmCode[m.pointInCode]
		op.doOp(m)
		if m.debugStack {
			println(m.outputStack())
		}
	}
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

func newVirtualMachine(wasmBytes []byte, storage []byte, config VMConfig) *Machine {
	machine := new(Machine)
	machine.pointInCode = 0
	machine.contractStorage = storage
	machine.debugStack = false
	machine.config = config
	machine.module = *decode(wasmBytes)
	machine.vmCode, machine.controlBlockStack = parseBytes(machine.module.codeSection[0].body)
	machine.locals = make([]uint64, len(machine.module.codeSection[0].localTypes))

	capacity := 20 * defaultPageSize
	machine.vmMemory = make([]byte, capacity)
	// Initialize empty memory.
	for i := 0; i < capacity; i++ {
		machine.vmMemory[i] = 0
	}

	return machine
}

func (m *Machine) popFromStack() uint64 {
	var ans uint64

	if m.debugStack {
		println("popping from stack")
	}
	ans, m.vmStack = m.vmStack[len(m.vmStack)-1], m.vmStack[:len(m.vmStack)-1]
	return ans
}
func (m *Machine) pushToStack(n uint64) {

	if m.debugStack {
		println("pushing to stack")
	}
	m.vmStack = append(m.vmStack, n)
}


// The caller of call2 has to pass in the function hash from the `inputData` field
// The function identifier will be the first 4 bytes of the data and the remaining 
// will be considered as function parameters


type GetCode func(hash []byte) (FunctionType, []OperationCommon, []ControlBlock)

func (m *Machine) call2(callBytes string, getCode GetCode) {

	// Structure: 0x[4 bytes func identifer][param1..][param2...][param3]
	// Note: The callbytes is following the wasm encoding scheme.

	bytes, err := hex.DecodeString(callBytes)

	if err != nil {
		panic("Unable to parse bytes for call2")
	}

	if len(bytes) < 4 {
		panic("Call2 - Invalid callbytes method identifier should be 4 bytes long")
	}

	funcIdentifier := bytes[:4]

	funcTypes, funcCode, controlStack := getCode(funcIdentifier)
	
	var params []uint64
	
	for i := uint64(len(funcIdentifier)); i < uint64(len(bytes)); i++ {

		valTypeByte := bytes[i]
		
		switch (valTypeByte) {
			case Op_i32_const:
				paramValue, count, err := DecodeInt32(reader(bytes[i+1:]))

				if (err != nil) {
					panic("Call2 - Error parsing function params i32")
				}

				params = append(params, uint64(paramValue))
				i += count

			case Op_i64_const:
				paramValue, count, err := DecodeInt64(reader(bytes[i+1:]))
				params = append(params, uint64(paramValue))

				if (err != nil) {
					panic("Call2 - Error parsing function params i64")
				}
				i += count

			case Op_f32_const:
				num := LE.Uint32(bytes[i+1: 4])
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

	if (expectedParamCount != incomingParamCount) {
		panic("Call2 - Param count mismatch")
	}

	// Maybe Check the types of each params if they matches signature?
	m.locals = params
	m.vmCode, m.controlBlockStack = funcCode, controlStack

	m.run()
}