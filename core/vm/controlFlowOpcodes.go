package vm
type Block struct {
	index uint32
}

func (op Block) doOp(m *Machine) {
	// Add some stack validation here
	stackLength  := len(m.vmStack)

	m.pointInCode++ // First skip this Block byte
	controlBlock := m.controlBlockStack[op.index]
	endOfThisBlock := End{controlBlock.index}

	for (m.vmCode[m.pointInCode] != endOfThisBlock) {
		m.vmCode[m.pointInCode].doOp(m)
	}

	finalStackLength := len(m.vmStack)

	if finalStackLength < stackLength {
		panic("Inconsistent stack after execution of Op_Block")
	}
}

type Br struct {
	index uint32
}

func (op Br) doOp(m *Machine) {

	if (len(m.controlBlockStack) < int(op.index)) {
		panic("Index where to branch out of range")
	}

	branch := m.controlBlockStack[op.index]
	endOfThisBlock := End{op.index}

	if (branch.op == Op_block || branch.op == Op_if) {
		// This means a break statement
		// @TODO(sdmg15) Replace this loop later with m.pointInCode += branch.endAt
		for (m.vmCode[m.pointInCode] != endOfThisBlock) {
			m.pointInCode += 1
		}
	} else if (branch.op == Op_loop) {
		// This means a continue statement
		m.pointInCode = branch.startAt
	} else {
		panic("Something went wrong while branching to index uint32(op.index)")
	}
}

type BrIf struct {
	index uint32
}

func (op BrIf) doOp(m *Machine) {
	if (len(m.controlBlockStack) < int(op.index)) {
		panic("Index where to branch out of range")
	}

	condition := uint32(m.popFromStack())

	if (condition != 0) {
		Br{op.index}.doOp(m)
	} else {
		NoOp{}.doOp(m)
	}
}

type If struct {
	controlBlock ControlBlock
}

func (op If) doOp(m *Machine) {
	condition := uint32(m.popFromStack())

	if (condition == 1) {
		// continue the execution
	} else {
		if (op.controlBlock.elseAt != 0) {
		} else {
			m.pointInCode += op.controlBlock.endAt + 1
		}
	}
}

type Loop struct {
	index uint32
}

func (op Loop) doOp(m *Machine) {
	stackLength  := len(m.vmStack)

	m.pointInCode++ // First skip this Loop byte
	controlBlock := m.controlBlockStack[op.index]
	endOfThisBlock := End{controlBlock.index}

	for (m.vmCode[m.pointInCode] != endOfThisBlock) {
		m.vmCode[m.pointInCode].doOp(m)
	}

	finalStackLength := len(m.vmStack)

	if finalStackLength < stackLength {
		panic("Inconsistent stack after execution of Op_Block")
	}

}

type NoOp struct {}

func (op NoOp) doOp(m *Machine) {
	m.pointInCode++
}

type UnReachable struct {}

func (op UnReachable) doOp(m *Machine) {
	m.pointInCode++
}

type End struct {
	index uint32
}

func (op End) doOp(m *Machine) {
	m.pointInCode++
}