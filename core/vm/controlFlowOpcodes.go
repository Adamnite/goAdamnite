package vm

type opIf struct {
	elsePoint int64
	endPoint  int64
}

func (op opIf) doOp(m *Machine) {
	//handling blocks as separate VMs
	if m.popFromStack() != uint64(0) {
		//do what they expect
		lastPoint := op.elsePoint
		if op.elsePoint == 0 {
			lastPoint = op.endPoint
		}
		vm := newVirtualMachine(m.vmCode[m.pointInCode+1:lastPoint], m.contractStorage, m.config)
		vm.vmStack = m.vmStack
		vm.debugStack = m.debugStack
		vm.run()
		m.vmStack = vm.vmStack

	} else if op.elsePoint != 0 {
		//do their else statement
		vm := newVirtualMachine(m.vmCode[op.elsePoint:op.endPoint], m.contractStorage, m.config)
		vm.vmStack = m.vmStack
		vm.debugStack = m.debugStack
		vm.run()
		m.vmStack = vm.vmStack
	}
	m.pointInCode = uint64(op.endPoint)
}

type noOp struct {}

func (op noOp) doOp(m *Machine) {
	m.pointInCode++
}

type unReachable struct {}

func (op unReachable) doOp(m *Machine) {
	m.pointInCode++
}