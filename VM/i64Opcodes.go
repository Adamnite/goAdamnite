package VM

import "math/bits"

type i64Const struct {
	val int64
	gas uint64
}

func (op i64Const) doOp(m *Machine) error {
	m.pushToStack(uint64(op.val))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Add struct {
	gas uint64
}

func (op i64Add) doOp(m *Machine) error {
	m.pushToStack(m.popFromStack() + m.popFromStack())
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Sub struct {
	gas uint64
}

func (op i64Sub) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()
	m.pushToStack(c1 - c2)
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Mul struct {
	gas uint64
}

func (op i64Mul) doOp(m *Machine) error {
	m.pushToStack(m.popFromStack() * m.popFromStack())
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Divs struct {
	gas uint64
}

func (op i64Divs) doOp(m *Machine) error {
	c2 := int64(m.popFromStack())
	c1 := int64(m.popFromStack())

	if c2 == 0 {
		panic("Division by zero")
	}

	m.pushToStack(uint64(c1 / c2))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Divu struct {
	gas uint64
}

func (op i64Divu) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()

	if c2 == 0 {
		panic("Division by zero")
	}

	m.pushToStack(uint64(c1 / c2))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Eqz struct {
	gas uint64
}

func (op i64Eqz) doOp(m *Machine) error {
	//i64.eqz is top value 0
	if len(m.vmStack) == 0 {
		m.pushToStack(1)
	} else if m.popFromStack() == 0 {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type i64Eq struct {
	gas uint64
}

func (op i64Eq) doOp(m *Machine) error {
	a := m.popFromStack() //just to prevent the compiler from thinking this is the same code on both sides
	if a == m.popFromStack() {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Ne struct {
	gas uint64
}

func (op i64Ne) doOp(m *Machine) error {
	a := m.popFromStack()
	if a == m.popFromStack() {
		m.pushToStack(0)
	} else {
		m.pushToStack(1)
	}
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64And struct {
	gas uint64
}

func (op i64And) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()

	m.pushToStack(c1 & c2)
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Or struct {
	gas uint64
}

func (op i64Or) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()

	m.pushToStack(c1 | c2)
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Xor struct {
	gas uint64
}

func (op i64Xor) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()

	m.pushToStack(c1 ^ c2)
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Les struct {
	gas uint64
}

func (op i64Les) doOp(m *Machine) error {
	c2 := int64(m.popFromStack())
	c1 := int64(m.popFromStack())
	if c1 <= c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Leu struct {
	gas uint64
}

func (op i64Leu) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()

	if c1 <= c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Ges struct {
	gas uint64
}

func (op i64Ges) doOp(m *Machine) error {
	c2 := int64(m.popFromStack())
	c1 := int64(m.popFromStack())
	if c1 >= c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Geu struct {
	gas uint64
}

func (op i64Geu) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()
	if c1 >= c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Lts struct {
	gas uint64
}

func (op i64Lts) doOp(m *Machine) error {
	c2 := int64(m.popFromStack())
	c1 := int64(m.popFromStack())
	if c1 < c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Ltu struct {
	gas uint64
}

func (op i64Ltu) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()
	if c1 < c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Gtu struct {
	gas uint64
}

func (op i64Gtu) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()
	if c1 > c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Gts struct {
	gas uint64
}

func (op i64Gts) doOp(m *Machine) error {
	c2 := int64(m.popFromStack())
	c1 := int64(m.popFromStack())
	if c1 > c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Shl struct {
	gas uint64
}

func (op i64Shl) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()
	m.pushToStack(c1 << (c2 % 64))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Shrs struct {
	gas uint64
}

func (op i64Shrs) doOp(m *Machine) error {
	c2 := int64(m.popFromStack())
	c1 := int64(m.popFromStack())

	m.pushToStack(uint64(c1 >> (c2 % 64)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Shru struct {
	gas uint64
}

func (op i64Shru) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()
	m.pushToStack(c1 >> (c2 % 64))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Rotl struct {
	gas uint64
}

func (op i64Rotl) doOp(m *Machine) error {
	c2 := int64(m.popFromStack())
	c1 := int64(m.popFromStack())
	m.pushToStack(bits.RotateLeft64(uint64(c1), int(c2)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Rotr struct {
	gas uint64
}

func (op i64Rotr) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()
	m.pushToStack(bits.RotateLeft64(uint64(c1), -1*int(c2)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Clz struct {
	gas uint64
}

func (op i64Clz) doOp(m *Machine) error {
	a := m.popFromStack()
	m.pushToStack(uint64(bits.LeadingZeros64(a)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Ctz struct {
	gas uint64
}

func (op i64Ctz) doOp(m *Machine) error {
	a := m.popFromStack()
	m.pushToStack(uint64(bits.TrailingZeros64(a)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64PopCnt struct {
	gas uint64
}

func (op i64PopCnt) doOp(m *Machine) error {
	a := m.popFromStack()
	m.pushToStack(uint64(bits.OnesCount64(a)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Rems struct {
	gas uint64
}

func (op i64Rems) doOp(m *Machine) error {
	c2 := m.popFromStack()
	c1 := m.popFromStack()

	if c2 == 0 {
		panic("integer division by zero")
	}

	m.pushToStack(c1 % c2)
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Remu struct {
	gas uint64
}

func (op i64Remu) doOp(m *Machine) error {
	c2 := int64(m.popFromStack())
	c1 := int64(m.popFromStack())

	if c2 == 0 {
		panic("integer division by zero")
	}

	m.pushToStack(uint64(c1 % c2))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}
