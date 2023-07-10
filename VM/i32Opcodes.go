package VM

import (
	"math"
	"math/bits"
)

type i32Sub struct {
	gas uint64
}

func (op i32Sub) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())
	m.pushToStack(c1 - c2)

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type i32Add struct {
	gas uint64
}

func (op i32Add) doOp(m *Machine) error {
	m.pushToStack(uint32(m.popFromStack()) + uint32(m.popFromStack()))

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type i32Mul struct {
	gas uint64
}

func (op i32Mul) doOp(m *Machine) error {
	m.pushToStack(uint64(uint32(uint32(m.popFromStack()) * uint32(m.popFromStack()))))

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Xor struct {
	gas uint64
}

func (op i32Xor) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	m.pushToStack(c1 ^ c2)
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Or struct {
	gas uint64
}

func (op i32Or) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	m.pushToStack(c1 | c2)
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type i32And struct {
	gas uint64
}

func (op i32And) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	m.pushToStack(uint64(uint32(c1 & c2)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type i32Remu struct {
	gas uint64
}

func (op i32Remu) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	if c2 != 0 {
		m.pushToStack(uint64(uint32(c1 % c2)))
	} else {
		panic("Division by Zero")
	}

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type i32Divu struct {
	gas uint64
}

func (op i32Divu) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	if c2 != 0 {
		m.pushToStack(uint64(uint32(c1 / c2)))
	} else {
		panic("Division by zero")
	}
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Const struct {
	val int32
	gas uint64
}

func (op i32Const) doOp(m *Machine) error {
	m.pushToStack(uint64(uint32(op.val)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Eqz struct {
	gas uint64
}

func (op i32Eqz) doOp(m *Machine) error {
	if len(m.vmStack) == 0 {
		m.pushToStack(1)
	} else if m.popFromStack() == 0 {
		m.pushToStack(uint64(uint32(1)))
	} else {
		m.pushToStack(uint64(uint32(0)))
	}

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type i32Eq struct {
	gas uint64
}

func (op i32Eq) doOp(m *Machine) error {
	if uint32(m.popFromStack()) == uint32(m.popFromStack()) {
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

type i32Ne struct {
	gas uint64
}

func (op i32Ne) doOp(m *Machine) error {
	if uint32(m.popFromStack()) != uint32(m.popFromStack()) {
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

type i32Lts struct {
	gas uint64
}

func (op i32Lts) doOp(m *Machine) error {
	c2 := int32(m.popFromStack())
	c1 := int32(m.popFromStack())

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

type i32Ltu struct {
	gas uint64
}

func (op i32Ltu) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

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

type i32Gtu struct {
	gas uint64
}

func (op i32Gtu) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

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

type i32Geu struct {
	gas uint64
}

func (op i32Geu) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

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

type i32Gts struct {
	gas uint64
}

func (op i32Gts) doOp(m *Machine) error {
	c2 := int32(m.popFromStack())
	c1 := int32(m.popFromStack())

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

type i32Ges struct {
	gas uint64
}

func (op i32Ges) doOp(m *Machine) error {
	c2 := int32(m.popFromStack())
	c1 := int32(m.popFromStack())

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

type i32Leu struct {
	gas uint64
}

func (op i32Leu) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

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

type i32Les struct {
	gas uint64
}

func (op i32Les) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

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

type i32Shl struct {
	gas uint64
}

func (op i32Shl) doOp(m *Machine) error {
	c2 := int32(m.popFromStack())
	c1 := int32(m.popFromStack())

	m.pushToStack(uint64(uint32(c1 << (c2 % 32))))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Shrs struct {
	gas uint64
}

func (op i32Shrs) doOp(m *Machine) error {
	c2 := int32(m.popFromStack())
	c1 := int32(m.popFromStack())

	m.pushToStack(uint64(int32(c1 >> (c2 % 32))))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Shru struct {
	gas uint64
}

func (op i32Shru) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	m.pushToStack(uint64(uint32(c1 >> (c2 % 32))))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Divs struct {
	gas uint64
}

func (op i32Divs) doOp(m *Machine) error {
	c2 := int32(uint32(m.popFromStack()))
	c1 := int32(uint32(m.popFromStack()))

	if c2 == 0 {
		panic("integer division by zero")
	}

	if c1 == math.MinInt32 && c2 == -1 {
		panic("signed integer overflow")
	}

	m.pushToStack(uint64(c1 / c2))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Rems struct {
	gas uint64
}

func (op i32Rems) doOp(m *Machine) error {
	c2 := int32(m.popFromStack())
	c1 := int32(m.popFromStack())

	if c2 == 0 {
		panic("integer division by zero")
	}

	m.pushToStack(uint64(uint32(c1 % c2)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Clz struct {
	gas uint64
}

func (op i32Clz) doOp(m *Machine) error {
	a := uint32(m.popFromStack())
	m.pushToStack(uint64(bits.LeadingZeros32(a)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Ctz struct {
	gas uint64
}

func (op i32Ctz) doOp(m *Machine) error {
	a := uint32(m.popFromStack())
	m.pushToStack(uint64(bits.TrailingZeros32(a)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32PopCnt struct {
	gas uint64
}

func (op i32PopCnt) doOp(m *Machine) error {
	a := uint32(m.popFromStack())
	m.pushToStack(uint64(bits.OnesCount32(a)))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Rotl struct {
	gas uint64
}

func (op i32Rotl) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	m.pushToStack(uint64(bits.RotateLeft32(c1, int(c2))))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Rotr struct {
	gas uint64
}

func (op i32Rotr) doOp(m *Machine) error {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	m.pushToStack(uint64(bits.RotateLeft32(c1, -int(c2))))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}
