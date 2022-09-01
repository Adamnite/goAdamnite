package vm

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
	params []int64 //any params that may be passed with the opcode?
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
	var op []operation
	return op
}

func popFromOpcodeStack(ops []operation) (operation, []operation) {
	return ops[len(ops)-1], ops[:len(ops)-1]
}
