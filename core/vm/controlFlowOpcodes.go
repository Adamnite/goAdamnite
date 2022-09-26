package vm

type opIf struct {
	startPoint uint64
	elsePoint uint64
	endPoint  uint64
	code []OperationCommon
	op byte
}

func (op opIf) doOp(m *Machine) {
	condition := m.popFromStack()

	if condition == uint64(1) {
		for i := 0; uint64(i) < uint64(len(op.code)); {
			op.code[i].doOp(m)
		}
		m.pointInCode += uint64(len(op.code))
	}

	if op.elsePoint != 0 {
		m.pointInCode += op.elsePoint // Move to program counter till the beginning of the else
	} else {
		// Execute go to end
	}
}

type noOp struct {}

func (op noOp) doOp(m *Machine) {
	m.pointInCode++
}

type unReachable struct {}

func (op unReachable) doOp(m *Machine) {
	m.pointInCode++
}

type opDrop struct {}

func (op opDrop) doOp(m *Machine) {
	m.popFromStack()
	m.pointInCode++
}

type opSelect struct {}

func (op opSelect) doOp(m *Machine) {
	a := uint32(m.popFromStack())
	b := uint32(m.popFromStack())
	c := uint32(m.popFromStack())

	if c != 0 {
		m.pushToStack(uint64(a))
	} else {
		m.pushToStack(uint64(b))
	}
	m.pointInCode++
}

type opElse struct {
	startPoint uint64
	endPoint uint64
	code []OperationCommon
}

func (op opElse) doOp(m *Machine) {

}
type opBrTable struct {
	labels []uint64
	defaultLabel uint64
}

func (op opBrTable) doOp(m *Machine) {
	label := m.popFromStack()
	if label > uint64(len(op.labels)) {
		print("Label is greater than the max label running the default lable")
		var op opBr
		op.doOp(m)
	}
	// Other code related to selecting the label from the `labels` vector
}
type opBr struct {
	// label uint64 // The label where to jump to
}

func (op opBr) doOp(m *Machine) {
	label := m.popFromStack()
	whereToGo := m.controlBlockStack[len(m.controlBlockStack) - int(label) - 1]
	// Execute the code of the place whereToGo
	// Note: Branching to a `block` is similar to `break`
	// Note: Branching to a `loop` is similar to `continue`

	for i := 0; uint64(i) < uint64(len(whereToGo.code)); {
		whereToGo.code[i].doOp(m)
	}

	m.pointInCode += uint64(len(whereToGo.code))
}

type opBrIf struct {}

func (op opBrIf) doOp(m *Machine) {
	condition := m.popFromStack()

	if condition == uint64(1) {
		var br opBr
		br.doOp(m) 
	} else {
		if (m.debugStack) {
			println("Condition not true, so continue")
		}
		m.pointInCode++
	}
}

type opEnd struct {}

func (op opEnd) doOp(m *Machine) {
}

type opBlock struct {
	startPoint uint64
	endPoint uint64
	code []OperationCommon
}

func (op opBlock) doOp(m *Machine) {
	var controlBlockElem controlBlock
	controlBlockElem.code = op.code
	controlBlockElem.startAt = op.startPoint
	controlBlockElem.endAt = op.endPoint
	m.controlBlockStack = append(m.controlBlockStack, controlBlockElem)
	m.pointInCode++
}


type opLoop struct {
	startPoint uint64
	endPoint uint64
	code []OperationCommon
}

func (op opLoop) doOp(m *Machine) {
	for i := 0; uint64(i) < uint64(len(op.code)); {
		op.code[i].doOp(m)
	}
	m.pointInCode += uint64(len(op.code))
}

type opReturn struct {}

func (op opReturn) doOp(m *Machine) {
	
}