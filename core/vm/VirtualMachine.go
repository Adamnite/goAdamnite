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
	contractStorage Storage  //the storage of the smart contracts data.
	vmMemory        []byte   //i believe the agreed on stack size was
	locals          []uint64 //local vals that the VM code can call
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
	opcode uint8
	params []uint8 //any params that may be passed with the opcode?
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
		ans += strconv.FormatUint(v, 16) + "\n"
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
	case 0x04: //if statement
		// if statements are passed a value type, but can ignore that and just assume any number would have the same response.
		// i.64 0 is 0x0000000000000000, i32 0 is 0x00000000, with a 64 bit stack, they equal the same thing.
		// this may be added so you could do cool things like only check half of a i.64 at a time, we will see.

		// if stack.pop !=0, do all steps until an else, or an end
		// if stack.pop ==0, go to the closest end, or else statement (whichever has a lower val)
		nextEnd := findNext(0x0B, m.pointInCode, m.vmCode)  //0x0B for end
		nextElse := findNext(0x05, m.pointInCode, m.vmCode) //0x04 for else

		if m.popFromStack() != uint64(0) {
			println("if case ran")
			completionPoint := nextElse
			if nextEnd < nextElse {
				completionPoint = nextEnd
			}
			m.pointInCode += 1 //run all the commands between this, and the completion point (the else statement, or the end for this if)
			for m.pointInCode < completionPoint {
				m.do(m.vmCode[m.pointInCode])
			}
		} else if nextElse < nextEnd && nextElse != 0 { //if there is an else statement
			println("else case ran")
			m.pointInCode = nextElse
			for m.pointInCode < nextEnd { //run all the commands between the else and the endPoint
				m.do(m.vmCode[m.pointInCode])
			}
		}
		m.pointInCode = nextEnd //no mater what, go to the end
	case 0x20: //local.get
		m.pushToStack(m.locals[op.params[0]]) //pushes the local value at index to the stack
	case 0x21: //local.set
		//im not too clear why this would be called.
		m.locals[op.params[0]] = m.popFromStack()
	case 0x42: //const i64
		// println("adding")
		if len(op.params) == 1 {
			m.pushToStack(uint64(op.params[0]))
		} else {
			m.pushToStack(uint64(0))
		}

		break
	case 0x50: //i64.eqz is top value 0
		// println("popping to check if is 0")
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
	// println("parsing apart the code")
	for _, s := range code {
		// println(s)
		values := strings.Split(s, " ")

		opcode, err := strconv.ParseUint(values[0], 16, 8)
		var params []uint8
		if err != nil {
			println("first error case")
			println(err.Error())
			//TODO: check if err is used
		}

		for i, a := range values {
			if i != 0 {
				param, err := strconv.ParseUint(a, 16, 8)
				if err != nil {
					println("second error case")
					println(err.Error())
					//TODO: check if err is used
				}
				params = append(params, uint8(param))
			}
		}
		operations = append(operations, operation{
			opcode: uint8(opcode),
			params: params})
	}
	return operations
}

func (m *Machine) popFromStack() uint64 {
	var ans uint64
	// println("popping from stack")
	ans, m.vmStack = m.vmStack[len(m.vmStack)-1], m.vmStack[:len(m.vmStack)-1]
	return ans
}
func (m *Machine) pushToStack(n uint64) {
	// println("pushing to stack")
	m.vmStack = append(m.vmStack, n)
}
func popFromOpcodeStack(ops []operation) (operation, []operation) {
	return ops[len(ops)-1], ops[:len(ops)-1]
}
func findNext(opcode uint8, start uint64, ops []operation) uint64 {
	for i := start; i < uint64(len(ops)); i++ {
		if ops[i].opcode == opcode {
			return i
		}
	}
	return 0 //means something went wrong
}
func findPrevious(opcode uint8, start uint64, ops []operation) uint64 {
	for i := start; i >= 0; i-- {
		if ops[i].opcode == opcode {
			return i
		}
	}
	return 0xFFFFFFFFFFFFFFFF //assume all 1's is an error.
}
