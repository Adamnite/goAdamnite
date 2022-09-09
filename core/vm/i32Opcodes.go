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

type i32And struct {}

func (op i32And) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() & m.popFromStack())
	m.pointInCode++
}

type i32Remu struct {} 

func (op i32Remu) doOp(m *Machine) {
	i1 := m.popFromStack()
	i2 := m.popFromStack()

	if i1 != 0 {
		m.pushToStack(i1 % i2)
	}

	m.pointInCode++
}


type i32Divu struct {}

func (op i32Divu) doOp(m *Machine) {
	i1 := m.popFromStack()
	i2 := m.popFromStack()

	if i1 != 0 {
		m.pushToStack(i1 / i2)
	}
	m.pointInCode++
}
