package VM

type Operationutils interface {
	doOp(m *Machine) error
}

type localGet struct {
	point int64
	gas   uint64
}

func (op localGet) doOp(m *Machine) error {
	if op.point == -1 { //use -1 to get it from the stack, since there cant be a negative index
		op.point = int64(m.popFromStack())
	}
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
	pointInStorage int64
	gas            uint64
}

func (op GlobalSet) doOp(m *Machine) error {
	if op.pointInStorage == -1 { //use -1 to get it from the stack, since there cant be a negative index
		op.pointInStorage = int64(m.popFromStack())
	}
	newValue := m.popFromStack()
	for len(m.contractStorage) <= int(op.pointInStorage) {
		m.contractStorage = append(m.contractStorage, 0)
	}
	m.contractStorage[op.pointInStorage] = newValue
	m.storageChanges[uint32(op.pointInStorage)] = newValue

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type GlobalGet struct {
	pointInStorage int64
	gas            uint64
}

func (op GlobalGet) doOp(m *Machine) error {
	if op.pointInStorage == -1 { //use -1 to get it from the stack, since there cant be a negative index
		op.pointInStorage = int64(m.popFromStack())
	}
	m.pushToStack(m.contractStorage[op.pointInStorage])

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}
