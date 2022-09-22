package vm

import "math"

type f64Const struct {
	val float64
}

func (op f64Const) doOp(m *Machine) {
	m.pushToStack(uint64(float64(op.val)))
	m.pointInCode++
}

type f64Eq struct {}

func (op f64Eq) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))

	if a == b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
}

type f64Ne struct {}

func (op f64Ne) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if a != b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
}

type f64Lt struct {}

func (op f64Lt) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if a < b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
}

type f64Gt struct {}

func (op f64Gt) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if a > b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
}

type f64Ge struct {}

func (op f64Ge) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if a >= b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
}

type f64Le struct {}

func (op f64Le) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if a <= b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
}

type f64Abs struct {}

func (op f64Abs) doOp(m *Machine) {
	val := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := math.Abs(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Neg struct {}

func (op f64Neg) doOp(m *Machine) {
	val := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := -val; c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Ceil struct {}

func (op f64Ceil) doOp(m *Machine) {
	val := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := math.Ceil(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Floor struct {}

func (op f64Floor) doOp(m *Machine) {
	val := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := math.Floor(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Trunc struct {}

func (op f64Trunc) doOp(m *Machine) {
	val := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := math.Trunc(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Nearest struct {}

func (op f64Nearest) doOp(m *Machine) {
	val := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := math.RoundToEven(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Sqrt struct {}

func (op f64Sqrt) doOp(m *Machine) {
	val := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := math.Sqrt(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Add struct {}

func (op f64Add) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := a + b; c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Sub struct {}

func (op f64Sub) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := a - b; c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Mul struct {}

func (op f64Mul) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := a * b; c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Div struct {}

func (op f64Div) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := a / b; c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Min struct {}

func (op f64Min) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := math.Min(a, b); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64Max struct {}

func (op f64Max) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := math.Max(a, b); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}

type f64CopySign struct {}

func (op f64CopySign) doOp(m *Machine) {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := math.Copysign(a, b); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
}