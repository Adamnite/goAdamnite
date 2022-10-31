package vm

import "math"

type i32Wrapi64 struct {}

func (op i32Wrapi64) doOp(m *Machine) {
	a := uint32(m.popFromStack())
	m.pushToStack(uint64(a))
	m.pointInCode++
}

type i32Truncsf32 struct {}

func (op i32Truncsf32) doOp(m *Machine) {
	a := math.Float32frombits(uint32(m.popFromStack()))

	if c := float32(math.Trunc(float64(a))); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(int32(c)))
	}
	m.pointInCode++
}


type i32Truncsf64 struct {}

func (op i32Truncsf64) doOp(m *Machine) {

	v := math.Float64frombits(uint64(m.popFromStack()))

	if c := math.Trunc(v); c != c {
		m.pushToStack(uint64(0x7FC00000))
	} else {
		m.pushToStack(uint64(int32(c)))
	}
	m.pointInCode++
}

type i64Extendsi32 struct {} 

func (op i64Extendsi32) doOp(m *Machine) {
	v := int32(m.popFromStack())
	m.pushToStack(uint64(v))
	m.pointInCode++
}


type i64Extendui32 struct {} 

func (op i64Extendui32) doOp(m *Machine) {
	v := uint32(m.popFromStack())
	m.pushToStack(uint64(v))
	m.pointInCode++
}


type i64Truncsf32 struct {}

func (op i64Truncsf32) doOp(m *Machine) {
	v := math.Float32frombits(uint32(m.popFromStack()))

	if c := math.Trunc(float64(v)); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(c))
	}
	m.pointInCode++
}

type i64Truncsf64 struct {}

func (op i64Truncsf64) doOp(m *Machine) {
	v := math.Float64frombits(uint64(m.popFromStack()))
	
	if c := math.Trunc(v); c != c {
		m.pushToStack(uint64(0x7FF8000000000001))
	} else {
		m.pushToStack(uint64(c))
	}
	m.pointInCode++
}

type f32Convertsi32 struct {}

func (op f32Convertsi32) doOp(m *Machine) {
	v := int32(m.popFromStack())
	m.pushToStack(uint64(math.Float32bits(float32(v))))
	m.pointInCode++
}

type f32Convertui32 struct {}

func (op f32Convertui32) doOp(m *Machine) {
	v := uint32(m.popFromStack())
	m.pushToStack(uint64(math.Float32bits(float32(v))))
	m.pointInCode++
}


type f32Convertsi64 struct {}

func (op f32Convertsi64) doOp(m *Machine) {
	v := m.popFromStack()
	m.pushToStack(uint64(math.Float32bits(float32(v))))
	m.pointInCode++
}

type f32Convertui64 struct {}

func (op f32Convertui64) doOp(m *Machine) {
	v := uint64(m.popFromStack())
	m.pushToStack(uint64(math.Float32bits(float32(v))))
	m.pointInCode++
}

type f32Demotef64 struct {}

func (op f32Demotef64) doOp(m *Machine) {
	v := math.Float64frombits(m.popFromStack())
	m.pushToStack(uint64(math.Float32bits(float32(v))))
	m.pointInCode++
}

type f64convertsi32 struct {}

func (op f64convertsi32) doOp(m *Machine) {
	v := int32(m.popFromStack())
	m.pushToStack(uint64(int32(math.Float64bits(float64(v)))))

	m.pointInCode++
}

type f64convertui32 struct {}

func (op f64convertui32) doOp(m *Machine) {
	v := uint32(m.popFromStack())
	m.pushToStack(uint64(uint32(math.Float64bits(float64(v)))))
	m.pointInCode++
}

type f64Convertsi64 struct {}

func (op f64Convertsi64) doOp(m *Machine) {
	v := int64(m.popFromStack())
	m.pushToStack(uint64(float64(v)))
	m.pointInCode++
}

type f64Convertui64 struct {}

func (op f64Convertui64) doOp(m *Machine) {
	v := m.popFromStack()
	m.pushToStack(uint64(math.Float64bits(float64(v))))
	m.pointInCode++
}

type f64Promotef32 struct {}

func (op f64Promotef32) doOp(m *Machine) {
	v := math.Float32frombits(uint32(m.popFromStack()))

	if c := float64(v); c == math.Float64frombits(0x7FF8000000000000) {
		m.pushToStack(0x7FF8000000000001)
	} else {
		m.pushToStack(uint64(math.Float64bits(c)))
	}
	m.pointInCode++;
}