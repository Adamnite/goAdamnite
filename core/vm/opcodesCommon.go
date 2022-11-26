package vm

type OperationCommon interface {
	doOp(m *Machine) error
}

type localGet struct {
	point int64
	gas   uint64
}

func (op localGet) doOp(m *Machine) error {
	m.pushToStack(m.locals[op.point]) //pushes the local value at index to the stack

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type localSet struct {
	point int64
	gas   uint64
}

func (op localSet) doOp(m *Machine) error {
	for len(m.locals) <= int(op.point) {
		m.locals = append(m.locals, uint64(0))
	}
	m.locals[op.point] = m.popFromStack()
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type currentMemory struct {
	gas uint64
}

func (op currentMemory) doOp(m *Machine) error {
	m.pushToStack(uint64(len(m.vmMemory))) //should be divided by 65536, or the page size constant.
	//this division can be handled by >>16

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type growMemory struct {
	gas uint64
}

func (op growMemory) doOp(m *Machine) error {
	amount := m.popFromStack()
	m.pushToStack(uint64(len(m.vmMemory))) //only should be pushed if it worked, but i don't see how this can't...
	for i := 0; i < int(amount); i++ {     //amount should be multiplied by 65536, or the Page Size Constant.
		// this constant value can be generated faster by <<16.
		m.vmMemory = append(m.vmMemory, byte(0))
	}

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type TeeLocal struct {
	val uint64
	gas uint64
}

func (op TeeLocal) doOp(m *Machine) error {
	v := m.popFromStack()
	m.pushToStack(uint64(v))
	m.pushToStack(uint64(v))
	localSet{int64(op.val), GasQuickStep}.doOp(m)

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type Drop struct {
	gas uint64
}

func (op Drop) doOp(m *Machine) error {
	m.popFromStack()
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type GlobalSet struct {
	pointInStorage uint32
	gas            uint64
}

func (op GlobalSet) doOp(m *Machine) error {
	m.pushToStack(m.contractStorage[op.pointInStorage])

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type GlobalGet struct {
	pointInStorage uint32
	gas            uint64
}

func (op GlobalGet) doOp(m *Machine) error {
	m.contractStorage[op.pointInStorage] = m.popFromStack()
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}
