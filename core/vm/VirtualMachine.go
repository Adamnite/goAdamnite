package vm

import (
	"encoding/binary"
	"fmt"
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
	run()

	do()
	outputStack() string
}

type ControlBlock struct {
	code []OperationCommon // Can this []byte instead ?
	startAt, elseAt, endAt uint64
	op byte // Contains the value of the opcode that triggered this
}
type Machine struct {
	VirtualMachine
	pointInCode     uint64
	module 			Module // The module that will be executed inside the VM
	vmCode          []OperationCommon
	vmStack         []uint64
	contractStorage Storage  //the storage of the smart contracts data.
	vmMemory        []byte   //i believe the agreed on stack size was
	locals          []uint64 //local vals that the VM code can call
	debugStack      bool     //should it output the stack every operation
	globals         []uint64
	controlBlockStack []ControlBlock // Represents the labels indexes at which br, br_if can jump to
	
	config			VMConfig
	gas             uint64 // The allocated gas for the code execution
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

type Storage struct {
	// the storage type, handling some standard things.

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

func newVirtualMachine(wasmBytes []byte, storage Storage, config VMConfig) *Machine {
	machine := new(Machine)
	machine.pointInCode = 0
	// machine.vmCode = parseCodeToOpcodes(code)
	// machine.vmCode = code
	machine.contractStorage = storage
	machine.debugStack = false
	machine.config = config
	machine.module = *decode(wasmBytes)

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
