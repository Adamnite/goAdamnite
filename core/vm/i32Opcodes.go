package vm

import (
	"math"
	"math/bits"
)

type i32Sub struct{}

func (op i32Sub) doOp(m *Machine) {
	b := uint32(uint32(m.popFromStack()))
	a := uint32(uint32(m.popFromStack()))
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

	if i2 != 0 {
		m.pushToStack(uint64(uint32(i1 % i2)))
	} else {
		panic("Division by Zero")
	}

	m.pointInCode++
}

type i32Divu struct{}

func (op i32Divu) doOp(m *Machine) {
	i1 := uint32(m.popFromStack())
	i2 := uint32(m.popFromStack())

	if i1 != 0 {
		m.pushToStack(uint64(uint32(i1 / i2)))
	} else {
		panic("Division by zero")
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


type i32Eqz struct {}

func (op i32Eqz) doOp(m *Machine) {
	if len(m.vmStack) == 0 {
		m.pushToStack(1)
	} else if m.popFromStack() == 0 {
		m.pushToStack(uint64(uint32(1)))
	} else {
		m.pushToStack(uint64(uint32(0)))
	}
	m.pointInCode++
}

type i32Eq struct {}

func (op i32Eq) doOp(m *Machine) {
	if uint32(m.popFromStack()) == uint32(m.popFromStack()) {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i32Ne struct {}

func (op i32Ne) doOp(m *Machine) {
	if uint32(m.popFromStack()) != uint32(m.popFromStack()) {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i32Lts struct {}

func (op i32Lts) doOp(m *Machine) {
	c2 := int32(m.popFromStack())
	c1 := int32(m.popFromStack())

	if c1 < c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i32Ltu struct {}

func (op i32Ltu) doOp(m *Machine) {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	if c1 < c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i32Gtu struct {}

func (op i32Gtu) doOp(m *Machine) {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	if c1 > c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i32Geu struct {}

func (op i32Geu) doOp(m *Machine) {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	if c1 >= c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i32Gts struct {}

func (op i32Gts) doOp(m *Machine) {
	c2 := int32(m.popFromStack())
	c1 := int32(m.popFromStack())

	if c1 > c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i32Ges struct {}

func (op i32Ges) doOp(m *Machine) {
	c2 := int32(m.popFromStack())
	c1 := int32(m.popFromStack())

	if c1 >= c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i32Leu struct {}

func (op i32Leu) doOp(m *Machine) {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	if c1 <= c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i32Les struct {}

func (op i32Les) doOp(m *Machine) {
	c2 := uint32(m.popFromStack())
	c1 := uint32(m.popFromStack())

	if c1 <= c2 {
		m.pushToStack(1)
	} else {
		m.pushToStack(0)
	}
	m.pointInCode++
}

type i32Shl struct {}

func (op i32Shl) doOp(m *Machine) {
	a := int32(m.popFromStack())
	b := uint32(m.popFromStack())

	m.pushToStack(uint64(uint32(a << (b % 32))))
	m.pointInCode++;
}

type i32Shrs struct {}

func (op i32Shrs) doOp(m *Machine) {
	a := int32(m.popFromStack())
	b := uint32(m.popFromStack())

	m.pushToStack(uint64(int32(a >> (b % 32))))
	m.pointInCode++;
}

type i32Shru struct {}

func (op i32Shru) doOp(m *Machine) {
	a := uint32(m.popFromStack())
	b := uint32(m.popFromStack())

	m.pushToStack(uint64(uint32(a >> (b % 32))))
	m.pointInCode++;
}

type i32Divs struct {}

func (op i32Divs) doOp(m *Machine) {
	a := int32(uint32(m.popFromStack()))
	b := int32(uint32(m.popFromStack()))

	if b == 0 {
		panic("integer division by zero")
	}

	if a == math.MinInt32 && b == -1 {
		panic("signed integer overflow")
	}

	m.pushToStack(uint64(a/b))
	m.pointInCode++
}

type i32Rems struct {}

func (op i32Rems) doOp(m *Machine) {
	a := int32(m.popFromStack())
	b := int32(m.popFromStack())

	if b == 0 {
		panic("integer division by zero")
	}
	
	m.pushToStack(uint64(uint32(a % b)))
	m.pointInCode++
}

type i32Clz struct {}

func (op i32Clz) doOp(m *Machine) {
	a := uint32(m.popFromStack())
	m.pushToStack(uint64(bits.LeadingZeros32(a)))
	m.pointInCode++
}

type i32Ctz struct {}

func (op i32Ctz) doOp(m *Machine) {
	a := uint32(m.popFromStack())
	m.pushToStack(uint64(bits.TrailingZeros32(a)))
	m.pointInCode++
}

type i32PopCnt struct {}

func (op i32PopCnt) doOp(m *Machine) {
	a := uint32(m.popFromStack())
	m.pushToStack(uint64(bits.OnesCount32(a)))
	m.pointInCode++
}


type i32Rotl struct {}

func (op i32Rotl) doOp(m *Machine) {
	a := uint32(m.popFromStack())
	b := uint32(m.popFromStack())

	m.pushToStack(uint64(bits.RotateLeft32(a, int(b))))
	m.pointInCode++
}


type i32Rotr struct {}

func (op i32Rotr) doOp(m *Machine) {
	a := uint32(m.popFromStack())
	b := uint32(m.popFromStack())

	m.pushToStack(uint64(bits.RotateLeft32(a, -int(b))))
	m.pointInCode++
}


