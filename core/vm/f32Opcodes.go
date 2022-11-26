package vm

import "math"

type f32Const struct {
	val float32
	gas uint64
}

func (op f32Const) doOp(m *Machine) error {
	m.pushToStack(uint64(float32(op.val)))

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type f32Eq struct {
	gas uint64
}

func (op f32Eq) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a == b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type f32Neq struct{
	gas uint64
}

func (op f32Neq) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a == b {
		m.pushToStack(uint64(0))
	} else {
		m.pushToStack(uint64(1))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type f32Lt struct{
	gas uint64
}

func (op f32Lt) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a < b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type f32Gt struct{
	gas uint64
}

func (op f32Gt) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a > b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type f32Ge struct{
	gas uint64
}

func (op f32Ge) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a >= b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type f32Le struct{
	gas uint64
}

func (op f32Le) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a <= b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type f32Abs struct{
	gas uint64
}

func (op f32Abs) doOp(m *Machine) error {
	val := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Abs(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type f32Neg struct{
	gas uint64
}

func (op f32Neg) doOp(m *Machine) error {
	val := math.Float32frombits(uint32(m.popFromStack()))

	if c := -val; c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type f32Ceil struct{
	gas uint64
}

func (op f32Ceil) doOp(m *Machine) error {
	val := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Ceil(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type f32Floor struct{
	gas uint64
}

func (op f32Floor) doOp(m *Machine) error {
	val := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Floor(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type f32Trunc struct{
	gas uint64
}

func (op f32Trunc) doOp(m *Machine) error {
	val := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Trunc(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type f32Nearest struct{
	gas uint64
}

func (op f32Nearest) doOp(m *Machine) error {
	val := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.RoundToEven(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type f32Sqrt struct{
	gas uint64
}

func (op f32Sqrt) doOp(m *Machine) error {
	val := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Sqrt(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type f32Add struct{
	gas uint64
}

func (op f32Add) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := a + b; c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type f32Sub struct{
	gas uint64
}

func (op f32Sub) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := a - b; c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type f32Mul struct{
	gas uint64
}

func (op f32Mul) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := a * b; c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type f32Div struct{
	gas uint64
}

func (op f32Div) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := a / b; c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type f32Max struct{
	gas uint64
}

func (op f32Max) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Max(float64(a), float64(b))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type f32Min struct{
	gas uint64
}

func (op f32Min) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Min(float64(a), float64(b))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type f32CopySign struct{
	gas uint64
}

func (op f32CopySign) doOp(m *Machine) error {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Copysign(float64(a), float64(b))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}
