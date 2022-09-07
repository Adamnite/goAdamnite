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
}
