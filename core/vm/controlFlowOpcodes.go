package vm
type Block struct {
	index uint32
}

func (op Block) doOp(m *Machine) {
	// Add some stack validation here
	stackLength := len(m.vmStack)

	m.pointInCode++ // First skip this Block byte
	controlBlock := m.controlBlockStack[op.index]
	endOfThisBlock := End{controlBlock.index}

	for (m.vmCode[m.pointInCode] != endOfThisBlock) {
		m.vmCode[m.pointInCode].doOp(m)
		m.pointInCode++
	}

	finalStackLength := len(m.vmStack)

	if stackLength < finalStackLength {
		panic("Inconsistent stack after execution of Op_Block")
	}
}

type Br struct {}

func (op Br) doOp(m *Machine) {

}

type BrIf struct {
}

func (op BrIf) doOp(m *Machine) {
	condition := uint32(m.popFromStack())

	if (condition == 1) {
		Br{}.doOp(m)
	} else {
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

}

func (op Loop) doOp(m *Machine) {

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