package vm

type OperationCommon interface {
	doOp(m *Machine)
}
type Operation struct{}

type localGet struct {
	point int64
}

func (op localGet) doOp(m *Machine) {
	m.pushToStack(m.locals[op.point]) //pushes the local value at index to the stack
	m.pointInCode++
}

type localSet struct {
	point int64
}

func (op localSet) doOp(m *Machine) {
	//im not too clear why this would be called.
	for len(m.locals) < int(op.point) {
		m.locals = append(m.locals, uint64(0))
	}
	m.locals[op.point] = m.popFromStack()
	m.pointInCode++
}

type currentMemory struct{}

func (op currentMemory) doOp(m *Machine) {
	m.pushToStack(uint64(len(m.vmMemory))) //should be divided by 65536, or the page size constant.
	//this division can be handled by >>16
	m.pointInCode++
}

type growMemory struct{}

func (op growMemory) doOp(m *Machine) {
	amount := m.popFromStack()
	m.pushToStack(uint64(len(m.vmMemory))) //only should be pushed if it worked, but i don't see how this can't...
	for i := 0; i < int(amount); i++ {     //amount should be multiplied by 65536, or the Page Size Constant.
		// this constant value can be generated faster by <<16.
		m.vmMemory = append(m.vmMemory, byte(0))
	}

	m.pointInCode++
}

type call struct {
	hashToCall []byte
}

func (op call) doOp(m *Machine) {
	var newVM *Machine
	if allZeros(op.hashToCall) {
		newVM = newVirtualMachine(m.vmCode, m.contractStorage)

	} else {
		newVM = newVirtualMachine(getCode(op.hashToCall), m.contractStorage)
	}
	newVM.vmStack = m.vmStack
	newVM.run()
}

func allZeros(arr []byte) bool {
	for _, v := range arr {
		if v != 0 {
			return false
		}
	}
	return true
}
