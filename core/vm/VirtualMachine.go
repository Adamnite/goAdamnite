package vm

import (
	"strconv"
	"strings"
)

type VirtualMachine interface {
	//functions that will be fully implemented later
	step()
	run()

	do()
}

type Machine struct {
	VirtualMachine
	//this needs stack, storage, and memory.
	vmStack         []operation //opcodes are stored as
	contractStorage Storage     //the storage of the smart contracts data.
	vmMemory        []int64     //i believe the agreed on stack size was
}

type Storage struct {
	// the storage type, handling some standard things.

}

type operation struct {
	//generate based on passed opcode list
	opcode int8
	params []int8 //any params that may be passed with the opcode?
}

func (m Machine) step() {
	if len(m.vmStack) > 0 {
		var op operation
		op, m.vmStack = popFromOpcodeStack(m.vmStack)
		m.do(op)
	}
}
func (m Machine) run() {
	for len(m.vmStack) > 0 {
		var op operation
		op, m.vmStack = popFromOpcodeStack(m.vmStack)
		m.do(op)
	}
}

func (m Machine) do(op operation) {
	//here is where the operations should run.
	//creating memory would look like
	// m.vmMemory = append(m.vmMemory, 0)
	switch op.opcode {
	case 0x00:

	}

}

func newVirtualMachine(code []string, storage Storage) *Machine {
	machine := new(Machine)
	machine.vmStack = parseCodeToOpcodes(code)
	machine.contractStorage = storage

	return machine
}
func parseCodeToOpcodes(code []string) []operation {
	// convert all the wasm data to opcodes with their function
	var operations []operation
	for _, s := range code {
		values := strings.Split(s, " ")

		opcode, err := strconv.ParseInt(values[0], 16, 8)
		var params []int8
		if err != nil {
			//TODO: check if err is used
		}

		for i, a := range values {
			if i != 0 {
				param, err := strconv.ParseInt(a, 16, 8)
				if err != nil {
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

func popFromOpcodeStack(ops []operation) (operation, []operation) {
	return ops[len(ops)-1], ops[:len(ops)-1]
}
