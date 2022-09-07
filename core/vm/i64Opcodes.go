package vm

func (op i64Add) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() + m.popFromStack())
	m.pointInCode++
}

type i64Sub struct{}

func (do i64Sub) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()
	m.pushToStack(b - a)
}

type i64Mul struct{}

func (do i64Mul) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() * m.popFromStack())
}

type i64Const struct {
	val int64
}

func (op i64Const) doOp(m *Machine) {
	m.pushToStack(uint64(op.val))
	m.pointInCode++
}

type i64Eqz struct{}

func (op i64Eqz) doOp(m *Machine) {
	//i64.eqz is top value 0
	if len(m.vmStack) == 0 {
		m.pushToStack(1)
	} else if m.popFromStack() == 0 {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
}
