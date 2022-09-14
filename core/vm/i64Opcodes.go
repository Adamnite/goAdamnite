package vm

import "math/bits"

type i64Const struct {
	val int64
}

func (op i64Const) doOp(m *Machine) {
	m.pushToStack(uint64(op.val))
	m.pointInCode++
}

type i64Add struct{}

func (op i64Add) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() + m.popFromStack())
	m.pointInCode++
}

type i64Sub struct{}

func (do i64Sub) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()
	m.pushToStack(b - a)
	m.pointInCode++
}

type i64Mul struct{}

func (do i64Mul) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() * m.popFromStack())
	m.pointInCode++
}

type i64Divs struct{}

func (do i64Divs) doOp(m *Machine) {
	a := int64(m.popFromStack())
	b := int64(m.popFromStack()) //b by a

	if a == 0 {
		panic("Division by 0")
	}

	m.pushToStack(uint64(b / a))
	m.pointInCode++
}

type i64Divu struct{}

func (do i64Divu) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack() //b by a

	if a == 0 {
		panic("Division by 0")
	}
	
	m.pushToStack(uint64(b / a))
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
	m.pointInCode++
}

type i64Eq struct{}

func (op i64Eq) doOp(m *Machine) {
	if m.popFromStack() == m.popFromStack() {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i64Ne struct{}

func (op i64Ne) doOp(m *Machine) {
	if m.popFromStack() == m.popFromStack() {
		m.pushToStack(0)
	} else {
		m.pushToStack(1)
	}
	m.pointInCode++
}

type i64And struct{}

func (op i64And) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() & m.popFromStack())
	m.pointInCode++
}

type i64Or struct{}

func (op i64Or) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() | m.popFromStack())
	m.pointInCode++
}

type i64Xor struct{}

func (op i64Xor) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() ^ m.popFromStack())
	m.pointInCode++
}

type i64Les struct{}

func (op i64Les) doOp(m *Machine) {
	a := int64(m.popFromStack())
	b := int64(m.popFromStack())
	if a <= b {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i64Leu struct{}

func (op i64Leu) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()
	if a <= b {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i64Ges struct{}

func (op i64Ges) doOp(m *Machine) {
	a := int64(m.popFromStack())
	b := int64(m.popFromStack())
	if a >= b {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i64Geu struct{}

func (op i64Geu) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()
	if a >= b {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i64Lts struct{}

func (op i64Lts) doOp(m *Machine) {
	a := int64(m.popFromStack())
	b := int64(m.popFromStack())
	if a < b {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i64Ltu struct{}

func (op i64Ltu) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()
	if a < b {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i64Gtu struct{}

func (op i64Gtu) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()
	if a > b {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i64Gts struct{}

func (op i64Gts) doOp(m *Machine) {
	a := int64(m.popFromStack())
	b := int64(m.popFromStack())
	if a > b {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}


type i64Shl struct{}

func (op i64Shl) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()
	m.pushToStack(b << a)
	m.pointInCode++
}

type i64Shrs struct{}

func (op i64Shrs) doOp(m *Machine) {
	a := int64(m.popFromStack())
	b := int64(m.popFromStack())
	m.pushToStack(uint64(b >> a))
	m.pointInCode++
}

type i64Shru struct{}

func (op i64Shru) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()
	m.pushToStack(b >> a)
	m.pointInCode++
}

type i64Rotl struct{}

func (op i64Rotl) doOp(m *Machine) {
	a := int64(m.popFromStack())
	b := int64(m.popFromStack())
	m.pushToStack(bits.RotateLeft64(uint64(b), int(a)))
	m.pointInCode++
}

type i64Rotr struct{}

func (op i64Rotr) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()
	m.pushToStack(bits.RotateLeft64(uint64(b), -1*int(a)))
	m.pointInCode++
}

type i64Clz struct {}

func (op i64Clz) doOp(m *Machine) {
	a := m.popFromStack()
	m.pushToStack(uint64(bits.LeadingZeros64(a)))
	m.pointInCode++
}

type i64Ctz struct {}

func (op i64Ctz) doOp(m *Machine) {
	a := m.popFromStack()
	m.pushToStack(uint64(bits.TrailingZeros64(a)))
	m.pointInCode++
}

type i64PopCnt struct {}

func (op i64PopCnt) doOp(m *Machine) {
	a := m.popFromStack()
	m.pushToStack(uint64(bits.OnesCount64(a)))
	m.pointInCode++
}

type i64Rems struct {}

func (op i64Rems) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()

	if b == 0 {
		panic("integer division by zero")
	}
	
	m.pushToStack(a % b)
	m.pointInCode++
}

type i64Remu struct {}

func (op i64Remu) doOp(m *Machine) {
	a := int64(m.popFromStack())
	b := int64(m.popFromStack())

	if b == 0 {
		panic("integer division by zero")
	}
	
	m.pushToStack(uint64(a % b))
	m.pointInCode++
}