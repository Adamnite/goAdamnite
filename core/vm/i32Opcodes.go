package vm

type i32Sub struct {}

func (op i32Sub) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()
	m.pushToStack(b - a)
	m.pointInCode++
}

type i32Add struct {}

func (op i32Add) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() + m.popFromStack())
	m.pointInCode++
}

type i32Mul struct {}

func (op i32Mul) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() * m.popFromStack())
	m.pointInCode++
}

type i32Xor struct {}

func (op i32Xor) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() ^ m.popFromStack())
	m.pointInCode++
}

type i32Or struct {}

func (op i32Or) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() | m.popFromStack())
	m.pointInCode++
}

