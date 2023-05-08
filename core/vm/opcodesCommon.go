package VM

type OperationCommon interface {
	doOp(m *Machine) error
}

type localGet struct {
	point int64
	gas   uint64
}

func (op localGet) doOp(m *Machine) error {
	m.pushToStack(m.locals[op.point]) //pushes the local value at index to the stack

	if !m.useAte(op.gas) {
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
	if !m.useAte(op.gas) {
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

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type Drop struct {
	gas uint64
}

func (op Drop) doOp(m *Machine) error {
	m.popFromStack()
	if !m.useAte(op.gas) {
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
	m.contractStorage[op.pointInStorage] = m.popFromStack()
	if !m.useAte(op.gas) {
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
	m.pushToStack(m.contractStorage[op.pointInStorage])

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}
