package vm

type i32Load struct {
	align uint32
	offset uint32
}

func (op i32Load) doOp(m *Machine) {

	// Take a memory immediate that contains an address offset and the expected 
	// alignment (expressed as the exponent of a power of 2)
	// https://webassembly.github.io/spec/core/syntax/instructions.html#memory-instructions
	ea := int(uint64(op.align) + uint64(op.offset))
	res := uint64(LE.Uint32(m.vmMemory[ea : ea + 4]))

	m.pushToStack(res)
	m.pointInCode++
}	

type i32Store struct {
	align uint32
	offset uint32
}

func (op i32Store) doOp(m *Machine) {
	value := uint32(m.popFromStack())
	ea := int(uint64(op.align) + uint64(op.offset))
	
	LE.PutUint32(m.vmMemory[ea : ea + 4], uint32(value))

	m.pointInCode++
}

type i64Load struct {
	align uint32
	offset uint32
}

func (op i64Load) doOp(m *Machine) {
	ea := int(uint64(op.align) + uint64(op.offset))
	value := (LE.Uint64(m.vmMemory[ea : ea + 8]))
	m.pushToStack(value)
	m.pointInCode++
}

type i64Store struct {
	align uint32
	offset uint32
}

func (op i64Store) doOp(m *Machine) {
	value := m.popFromStack()
	ea := int(uint64(op.align) + uint64(op.offset))
	LE.PutUint64(m.vmMemory[ea : ea + 8], uint64(value))
	m.pointInCode++
}

type i32Load8s struct {
	align uint32
	offset uint32
}

func (op i32Load8s) doOp(m *Machine) {
	ea := int(uint64(op.align) + uint64(op.offset))
	value := uint64(int8(m.vmMemory[ea]))
	m.pushToStack(value)
}

type i32Store8 struct {
	align uint32
	offset uint32
}

func (op i32Store8) doOp(m *Machine) {
	value := uint32(m.popFromStack())
	ea := int(uint64(op.align) + uint64(op.offset))

	m.vmMemory[ea] = byte(value)
	m.pointInCode++
}

type i32Load8u struct {
	align uint32
	offset uint32
}

func (op i32Load8u) doOp(m *Machine) {
	ea := int(uint64(op.align) + uint64(op.offset))
	res := int64(m.vmMemory[ea])

	m.pushToStack(uint64(res))
	m.pointInCode++
}

type i64Load16s struct {
	align uint32
	offset uint32
}

func (op i64Load16s) doOp(m *Machine) {
	ea := int(uint64(op.align) + uint64(op.offset))
	res := int64(int16(LE.Uint16(m.vmMemory[ea : ea + 2])))
	m.pushToStack(uint64(res))
	m.pointInCode++
}

type i32Load16u struct {
	align uint32
	offset uint32
}

func (op i32Load16u) doOp(m *Machine) {
	ea := int(uint64(op.align) + uint64(op.offset))
	res := uint64(int16(LE.Uint16(m.vmMemory[ea : ea + 2])))
	m.pushToStack(res)
	m.pointInCode++
}

type i64Load32s struct {
	align uint32
	offset uint32
}

func (op i64Load32s) doOp(m *Machine) {
	ea := int(uint64(op.align) + uint64(op.offset))
	res := int64(int32(LE.Uint32(m.vmMemory[ea : ea + 4])))
	m.pushToStack(uint64(res))
	m.pointInCode++
}


type i32Store16 struct {
	align uint32
	offset uint32
}

func (op i32Store16) doOp(m *Machine) {
	ea := int(uint64(op.align) + uint64(op.offset))
	value := uint32(m.popFromStack())
	LE.PutUint16(m.vmMemory[ea : ea + 2], uint16(value))
	m.pointInCode++
}
