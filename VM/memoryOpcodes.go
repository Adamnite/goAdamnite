package VM

type currentMemory struct {
	gas uint64
}

func (op currentMemory) doOp(m *Machine) error {
	m.pushToStack(uint64(len(m.vmMemory))) //should be divided by 65536, or the page size constant.
	//this division can be handled by >>16

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type growMemory struct {
	gas uint64
}

func (op growMemory) doOp(m *Machine) error {
	amount := m.popFromStack()
	m.pushToStack(uint64(len(m.vmMemory))) //only should be pushed if it worked, but i don't see how this can't...
	for i := 0; i < int(amount); i++ {     //amount should be multiplied by 65536, or the Page Size Constant.
		// this constant value can be generated faster by <<16.
		m.vmMemory = append(m.vmMemory, byte(0))
	}

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Load struct {
	align  uint32
	offset uint32
	gas    uint64
}

func (op i32Load) doOp(m *Machine) error {

	// Take a memory immediate that contains an address offset and the expected
	// alignment (expressed as the exponent of a power of 2)
	// https://webassembly.github.io/spec/core/syntax/instructions.html#memory-instructions
	index := uint64(m.popFromStack())
	ea := int(index + uint64(op.offset))
	res := uint64(LE.Uint32(m.vmMemory[ea : ea+4]))

	m.pushToStack(res)

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type i32Store struct {
	align  uint32
	offset uint32
	gas    uint64
}

// @TODO if the offset is greater than the prev
// allocated one we compute the required ate cost for expansion

func (op i32Store) doOp(m *Machine) error {
	value := uint32(m.popFromStack())
	index := uint32(m.popFromStack())
	ea := int(uint64(index) + uint64(op.offset))

	LE.PutUint32(m.vmMemory[ea:ea+4], uint32(value))

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type i64Load struct {
	align  uint32
	offset uint32
	gas    uint64
}

func (op i64Load) doOp(m *Machine) error {
	index := uint64(m.popFromStack())

	ea := int(index + uint64(op.offset))
	value := (LE.Uint64(m.vmMemory[ea : ea+8]))
	m.pushToStack(value)
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type i64Store struct {
	align  uint32
	offset uint32
	gas    uint64
}

func (op i64Store) doOp(m *Machine) error {
	value := uint32(m.popFromStack())
	index := uint32(m.popFromStack())
	ea := int(uint64(index) + uint64(op.offset))
	LE.PutUint64(m.vmMemory[ea:ea+8], uint64(value))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Load8s struct {
	align  uint32
	offset uint32
	gas    uint64
}

func (op i32Load8s) doOp(m *Machine) error {
	index := uint64(m.popFromStack())
	ea := int(index + uint64(op.offset))
	value := uint64(int8(m.vmMemory[ea]))
	m.pushToStack(value)

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type i32Store8 struct {
	align  uint32
	offset uint32
	gas    uint64
}

func (op i32Store8) doOp(m *Machine) error {
	value := uint32(m.popFromStack())
	index := uint32(m.popFromStack())
	ea := int(uint64(index) + uint64(op.offset))

	m.vmMemory[ea] = byte(value)
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Load8u struct {
	align  uint32
	offset uint32
	gas    uint64
}

func (op i32Load8u) doOp(m *Machine) error {
	index := uint64(m.popFromStack())
	ea := int(index + uint64(op.offset))
	res := int64(m.vmMemory[ea])

	m.pushToStack(uint64(res))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Load16s struct {
	align  uint32
	offset uint32
	gas    uint64
}

func (op i64Load16s) doOp(m *Machine) error {
	index := uint64(m.popFromStack())
	ea := int(index + uint64(op.offset))
	res := int64(int16(LE.Uint16(m.vmMemory[ea : ea+2])))
	m.pushToStack(uint64(res))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Load16u struct {
	align  uint32
	offset uint32
	gas    uint64
}

func (op i32Load16u) doOp(m *Machine) error {
	index := uint64(m.popFromStack())

	ea := int(index + uint64(op.offset))
	res := uint64(int16(LE.Uint16(m.vmMemory[ea : ea+2])))
	m.pushToStack(res)
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i64Load32s struct {
	align  uint32
	offset uint32
	gas    uint64
}

func (op i64Load32s) doOp(m *Machine) error {
	index := uint64(m.popFromStack())
	ea := int(index + uint64(op.offset))
	res := int64(int32(LE.Uint32(m.vmMemory[ea : ea+4])))
	m.pushToStack(uint64(res))
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	m.pointInCode++
	return nil
}

type i32Store16 struct {
	align  uint32
	offset uint32
	gas    uint64
}

func (op i32Store16) doOp(m *Machine) error {
	value := uint32(m.popFromStack())
	index := uint32(m.popFromStack())
	ea := int(uint64(index) + uint64(op.offset))
	LE.PutUint16(m.vmMemory[ea:ea+2], uint16(value))

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}
