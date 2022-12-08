package vm

import "math"

type f64Const struct {
	val float64
	gas uint64
}

func (op f64Const) doOp(m *Machine) error {
	m.pushToStack(uint64(float64(op.val)))
	m.pointInCode++
	return nil
}

type f64Eq struct{
	gas uint64

}

func (op f64Eq) doOp(m *Machine) error {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))

	if a == b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
	return nil
}

type f64Ne struct{
	gas uint64

}

func (op f64Ne) doOp(m *Machine) error {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))

	if a != b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
	return nil
}

type f64Lt struct{
	gas uint64

}

func (op f64Lt) doOp(m *Machine) error {
	b := uint64(m.popFromStack())
	a := uint64(m.popFromStack())

	if a < b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
	return nil
}

type f64Gt struct{
	gas uint64

}

func (op f64Gt) doOp(m *Machine) error {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))

	if a > b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
	return nil
}

type f64Ge struct{
	gas uint64

}

func (op f64Ge) doOp(m *Machine) error {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))

	if a >= b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
	return nil
}

type f64Le struct{
	gas uint64

}

func (op f64Le) doOp(m *Machine) error {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))

	if a <= b {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
	m.pointInCode++
	return nil
}

type f64Abs struct{
	gas uint64

}

func (op f64Abs) doOp(m *Machine) error {
	val := math.Float64frombits(uint64(m.popFromStack()))

	if c := math.Abs(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}

type f64Neg struct{
	gas uint64

}

func (op f64Neg) doOp(m *Machine) error {
	val := math.Float64frombits(uint64(m.popFromStack()))

	if c := -val; c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}

type f64Ceil struct{
	gas uint64

}

func (op f64Ceil) doOp(m *Machine) error {
	val := math.Float64frombits(uint64(m.popFromStack()))

	if c := math.Ceil(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}

type f64Floor struct{
	gas uint64

}

func (op f64Floor) doOp(m *Machine) error {
	val := math.Float64frombits(uint64(m.popFromStack()))

	if c := math.Floor(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}

type f64Trunc struct{
	gas uint64

}

func (op f64Trunc) doOp(m *Machine) error {
	val := math.Float64frombits(uint64(m.popFromStack()))

	if c := math.Trunc(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}

type f64Nearest struct{
	gas uint64

}

func (op f64Nearest) doOp(m *Machine) error {
	val := math.Float64frombits(uint64(m.popFromStack()))

	if c := math.RoundToEven(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}

type f64Sqrt struct{
	gas uint64

}

func (op f64Sqrt) doOp(m *Machine) error {
	val := math.Float64frombits(uint64(m.popFromStack()))

	if c := math.Sqrt(val); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}

type f64Add struct{
	gas uint64

}

func (op f64Add) doOp(m *Machine) error {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))

	if c := a + b; c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}

type f64Sub struct{
	gas uint64

}

func (op f64Sub) doOp(m *Machine) error {
	b := uint64(float64(m.popFromStack()))
	a := uint64(float64(m.popFromStack()))

	m.pushToStack(uint64(a-b))
	m.pointInCode++
	return nil
}

type f64Mul struct{
	gas uint64

}

func (op f64Mul) doOp(m *Machine) error {
	a := uint64(m.popFromStack())
	b := uint64(m.popFromStack())

	m.pushToStack(uint64(a*b))
	m.pointInCode++
	return nil
}

type f64Div struct{
	gas uint64

}

func (op f64Div) doOp(m *Machine) error {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))

	if c := a / b; c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}

type f64Min struct{
	gas uint64

}

func (op f64Min) doOp(m *Machine) error {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))

	if c := math.Min(a, b); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}

type f64Max struct{
	gas uint64

}

func (op f64Max) doOp(m *Machine) error {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))

	if c := math.Max(a, b); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}

type f64CopySign struct{
	gas uint64

}

func (op f64CopySign) doOp(m *Machine) error {
	a := math.Float64frombits(uint64(m.popFromStack()))
	b := math.Float64frombits(uint64(m.popFromStack()))

	if c := math.Copysign(a, b); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++
	return nil
}
