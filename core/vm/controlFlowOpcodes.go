package vm
type Block struct {
	index uint32 // The index of its controlblock inside the controlBlockStack
}

func (op Block) doOp(m *Machine) {
	// Add some stack validation here
	stackLength  := len(m.vmStack)

	m.pointInCode++ // First skip this Block byte
	control := m.controlBlockStack[op.index]

	for (m.pointInCode != control.endAt) {
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

	// An important note about the labels is that the innermost one has the index 0 and the outtermost one has the index N
	// https://webassembly.github.io/spec/core/bikeshed/index.html#control-instructions%E2%91%A0
	if (len(m.controlBlockStack) < int(op.index)) {
		panic("Index where to branch out of range")
	}

	branch := m.controlBlockStack[len(m.controlBlockStack) - int(op.index) - 1]

	if (branch.op == Op_block || branch.op == Op_if) {
		// This means a break statement
		m.pointInCode = branch.endAt
	} else if (branch.op == Op_loop) {
		// This means a continue statement
		m.pointInCode = branch.startAt + 1 // +1 To skip the block byte
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

	// Once the pointInCode becomes bigger than the endAt then it means we branched to a block
	for (m.pointInCode < controlBlock.endAt) {
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