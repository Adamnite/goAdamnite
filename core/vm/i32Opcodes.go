package vm

type i32Sub struct{}

func (op i32Sub) doOp(m *Machine) {
	a := uint32(uint32(m.popFromStack()))
	b := uint32(uint32(m.popFromStack()))
	m.pushToStack(uint64(uint32(b - a)))
	m.pointInCode++
}

type i32Add struct{}

func (op i32Add) doOp(m *Machine) {
	m.pushToStack(uint64(uint32(uint32(m.popFromStack()) + uint32(m.popFromStack()))))
	m.pointInCode++
}

type i32Mul struct{}

func (op i32Mul) doOp(m *Machine) {
	m.pushToStack(uint64(uint32(uint32(m.popFromStack()) * uint32(m.popFromStack()))))
	m.pointInCode++
}

type i32Xor struct{}

func (op i32Xor) doOp(m *Machine) {
	m.pushToStack(uint64(uint32(uint32(m.popFromStack()) ^ uint32(m.popFromStack()))))
	m.pointInCode++
}

type i32Or struct{}

func (op i32Or) doOp(m *Machine) {
	m.pushToStack(uint64(uint32(uint32(m.popFromStack()) | uint32(m.popFromStack()))))
	m.pointInCode++
}

type i32And struct{}

func (op i32And) doOp(m *Machine) {
	m.pushToStack(uint64(uint32(uint32(m.popFromStack()) & uint32(m.popFromStack()))))
	m.pointInCode++
}

type i32Remu struct{}

func (op i32Remu) doOp(m *Machine) {
	i1 := uint32(m.popFromStack())
	i2 := uint32(m.popFromStack())

	if i1 != 0 {
		m.pushToStack(uint64(uint32(i1 % i2)))
	}

	m.pointInCode++
}

type i32Divu struct{}

func (op i32Divu) doOp(m *Machine) {
	i1 := uint32(m.popFromStack())
	i2 := uint32(m.popFromStack())

	if i1 != 0 {
		m.pushToStack(uint64(uint32(i1 / i2)))
	}
	m.pointInCode++
}

type i32Const struct {
	val int32
}

func (op i32Const) doOp(m *Machine) {
	m.pushToStack(uint64(uint32(op.val)))
	m.pointInCode++
}
