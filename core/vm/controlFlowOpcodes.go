package vm
type Block struct {
	index uint32 // The index of its controlblock inside the controlBlockStack
}

func (op Block) doOp(m *Machine) {
	// Add some stack validation here
	stackLength  := len(m.vmStack)

	m.pointInCode++ // First skip this Block byte
	control := m.controlBlockStack[op.index]

	for (m.pointInCode < control.endAt) {
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
	index uint32 // The index of its controlblock inside the controlBlockStack
}

func (op If) doOp(m *Machine) {
	// @TODO(sdmg15) Check if the top of the stack is the same type as signature in If/Else/End
	condition := uint32(m.popFromStack())

	stackLen := len(m.vmStack)
	
	controlBlock := m.controlBlockStack[int(op.index)]

	if (controlBlock.op != Op_if) {
		panic("Invalid Operand retrieved from stack - Op_If expected")
	}
	
	if (condition != 0) {
		end := controlBlock.endAt

		if (controlBlock.elseAt != 0) {
			end = controlBlock.elseAt - 1 //
		}

		for (m.pointInCode != end) {
			m.vmCode[m.pointInCode].doOp(m)
		}

		if (controlBlock.elseAt != 0) {
			m.pointInCode = controlBlock.endAt
		}

	} else if (controlBlock.elseAt != 0) {
		m.pointInCode = controlBlock.elseAt + 1 // + 1 to skip the block byte
	} else {
		m.pointInCode = controlBlock.endAt
	}

	if (len(m.vmStack) < stackLen) {
		panic("Inconsistent stack after execution of Op_If")
	}

}

type Else struct {
	index uint32 // The index of its controlblock inside the controlBlockStack should have the same value as the If one
}

func (op Else) doOp(m *Machine) {

	stackLen := len(m.vmStack)
	controlBlock := m.controlBlockStack[len(m.controlBlockStack) - 1]

	for (m.pointInCode != controlBlock.endAt) {
		m.vmCode[m.pointInCode].doOp(m)
	}

	if (len(m.vmStack) < stackLen) {
		panic("Inconsistent stack after execution of Op_Else")
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

type Return struct {}

func (op Return) doOp(m *Machine) {
	// branch := m.controlBlockStack[0]

	if (len(m.vmStack) > 0) {
		res := m.popFromStack()

		for len(m.vmStack) != 0 {
			m.popFromStack()
			m.pointInCode++
		}
		m.pushToStack(uint64(res))
	} else {
		m.pointInCode += uint64(len(m.vmCode) - 2) // -1 for range and -1 for staying at End{} of function
	}
}

type Call struct {
	funcIndex uint32
}

func (op Call) doOp(m *Machine) {

	if int(op.funcIndex) >= len(m.module.typeSection[op.funcIndex].params) {
		panic("invalid function index")
	}

	// Pop the required params from stack
	params := m.module.typeSection[op.funcIndex].params
	poppedParams := []uint64{}
	for i := len(params); i != 0; i--{
		poppedParams = append(poppedParams, m.popFromStack())
	}
	code := m.module.codeSection[op.funcIndex].body
	
	m = newVirtualMachine(code, m.contractStorage, &m.config, 1000)
	m.locals = poppedParams
	m.run()

}

type CallIndirect struct {}

func (op CallIndirect) doOp(m *Machine) {
}