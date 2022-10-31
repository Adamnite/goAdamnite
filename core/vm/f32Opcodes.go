package vm

import "math"

type f32Const struct {
	val float32
}

func (op f32Const) doOp(m *Machine) {
	m.pushToStack(uint64(float32(op.val)))
	m.pointInCode++
}

type f32Eq struct {}

func (op f32Eq) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a == b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
}

type f32Neq struct {}

func (op f32Neq) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a == b {
		m.pushToStack(uint64(0))
	} else {
		m.pushToStack(uint64(1))
	}
	m.pointInCode++
}

type f32Lt struct {}

func (op f32Lt) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a < b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
}


type f32Gt struct {}

func (op f32Gt) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a > b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
}

type f32Ge struct {}

func (op f32Ge) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a >= b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
}

type f32Le struct {}

func (op f32Le) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if a <= b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
}

type f32Abs struct {}

func (op f32Abs) doOp(m *Machine) {
	val := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Abs(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32Neg struct {}

func (op f32Neg) doOp(m *Machine) {
	val := math.Float32frombits(uint32(m.popFromStack()))
	
	if c := -val; c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32Ceil struct {}

func (op f32Ceil) doOp(m *Machine) {
	val := math.Float32frombits(uint32(m.popFromStack()))
	
	if c := float32(math.Ceil(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32Floor struct {}

func (op f32Floor) doOp(m *Machine) {
	val := math.Float32frombits(uint32(m.popFromStack()))
	
	if c := float32(math.Floor(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32Trunc struct {}

func (op f32Trunc) doOp(m *Machine) {
	val := math.Float32frombits(uint32(m.popFromStack()))
	
	if c := float32(math.Trunc(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32Nearest struct {}

func (op f32Nearest) doOp(m *Machine) {
	val := math.Float32frombits(uint32(m.popFromStack()))
	
	if c := float32(math.RoundToEven(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32Sqrt struct {}

func (op f32Sqrt) doOp(m *Machine) {
	val := math.Float32frombits(uint32(m.popFromStack()))
	
	if c := float32(math.Sqrt(float64(val))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32Add struct {}

func (op f32Add) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := a + b; c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32Sub struct {}

func (op f32Sub) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := a - b; c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32Mul struct {}

func (op f32Mul) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := a * b; c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32Div struct {}

func (op f32Div) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := a / b; c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32Max struct {}

func (op f32Max) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Max(float64(a), float64(b))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}


type f32Min struct {}

func (op f32Min) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Min(float64(a), float64(b))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

type f32CopySign struct {}

func (op f32CopySign) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))
	b := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Copysign(float64(a), float64(b))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(math.Float32bits(c)))
	}
	m.pointInCode++
}

