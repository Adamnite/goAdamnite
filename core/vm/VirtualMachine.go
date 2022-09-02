package vm

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

var LE = binary.LittleEndian //an easier way to call little endian. I personally am not the biggest fan of LE,
// however, it is the specified standard of WASM.
// even though web applications traditionally used BE
type VirtualMachine interface {
	//functions that will be fully implemented later
	step()
	run()

	do()
	outputStack() string
}

type Machine struct {
	VirtualMachine
	pointInCode     uint64
	vmCode          []operation //opcodes are stored as
	vmStack         []uint64
	contractStorage Storage //the storage of the smart contracts data.
	vmMemory        []byte  //i believe the agreed on stack size was
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

type operation struct {
	//generate based on passed opcode list
	opcode int8
	params []int8 //any params that may be passed with the opcode?
}

func (m *Machine) step() {
	if m.pointInCode < uint64(len(m.vmCode)) {
		op := m.vmCode[m.pointInCode]
		m.do(op)
	}
}
func (m *Machine) run() {
	for m.pointInCode < uint64(len(m.vmCode)) {
		op := m.vmCode[m.pointInCode]
		m.do(op)
	}
}
func (m *Machine) outputStack() string {
	ans := ""
	for _, v := range m.vmStack {
		ans += fmt.Sprint(v) + "\n"
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

func (m *Machine) do(op operation) {
	//here is where the operations should run.
	//creating memory would look like
	// m.vmMemory = append(m.vmMemory, 0)
	switch op.opcode {
	case 0x42: //const i64
		println("adding")
		if len(op.params) == 1 {
			m.pushToStack(uint64(op.params[0]))
		} else {
			m.pushToStack(uint64(0))
		}

		break
	case 0x50: //i64.eqz is top value 0
		println("popping to check if is 0")
		if len(m.vmStack) == 0 {
			m.pushToStack(1)
		} else if m.popFromStack() == 0 {
			m.pushToStack(uint64(1))
		} else {
			m.pushToStack(uint64(0))
		}
		break
	case 0x00: //unreachable
	case 0x01: //nop
	default:
		break
	}
	m.pointInCode += 1

}

func newVirtualMachine(code []string, storage Storage) *Machine {
	machine := new(Machine)
	machine.pointInCode = 0
	machine.vmCode = parseCodeToOpcodes(code)
	machine.contractStorage = storage

	return machine
}
func parseCodeToOpcodes(code []string) []operation {
	// convert all the wasm data to opcodes with their function
	var operations []operation
	println("parsing apart the code")
	for _, s := range code {
		println(s)
		values := strings.Split(s, " ")

		opcode, err := strconv.ParseInt(values[0], 16, 8)
		var params []int8
		if err != nil {
			println("first error case")
			println(err.Error())
			//TODO: check if err is used
		}

		for i, a := range values {
			if i != 0 {
				param, err := strconv.ParseInt(a, 16, 8)
				if err != nil {
					println("second error case")
					println(err)
					//TODO: check if err is used
				}
				params = append(params, int8(param))
			}
		}
		operations = append(operations, operation{
			opcode: int8(opcode),
			params: params})
	}
	return operations
}

func (m *Machine) popFromStack() uint64 {
	var ans uint64
	println("popping from stack")
	ans, m.vmStack = m.vmStack[len(m.vmStack)-1], m.vmStack[:len(m.vmStack)-1]
	return ans
}
func (m *Machine) pushToStack(n uint64) {
	println("pushing to stack")
	m.vmStack = append(m.vmStack, n)
}
func popFromOpcodeStack(ops []operation) (operation, []operation) {
	return ops[len(ops)-1], ops[:len(ops)-1]
}
