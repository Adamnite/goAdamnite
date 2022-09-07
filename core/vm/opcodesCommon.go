package vm

type OperationCommon interface {
	doOp(m *Machine)
}
type Operation struct{}
type i64Add struct{}

func (op i64Add) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() + m.popFromStack())
	m.pointInCode++
}

type i64Sub struct{}

func (do i64Sub) doOp(m *Machine) {
	a := m.popFromStack()
	b := m.popFromStack()
	m.pushToStack(b - a)
}

type i64Mul struct{}

func (do i64Mul) doOp(m *Machine) {
	m.pushToStack(m.popFromStack() * m.popFromStack())
}

type i64Const struct {
	val int64
}

func (op i64Const) doOp(m *Machine) {
	m.pushToStack(uint64(op.val))
	m.pointInCode++
}

type i64Eqz struct{}

func (op i64Eqz) doOp(m *Machine) {
	//i64.eqz is top value 0
	if len(m.vmStack) == 0 {
		m.pushToStack(1)
	} else if m.popFromStack() == 0 {
		m.pushToStack(uint64(1))
	} else {
		m.pushToStack(uint64(0))
	}
}

type localGet struct {
	point int64
}

func (op localGet) doOp(m *Machine) {
	m.pushToStack(m.locals[op.point]) //pushes the local value at index to the stack
	m.pointInCode++
}

type localSet struct {
	point int64
}

func (op localSet) doOp(m *Machine) {
	//im not too clear why this would be called.
	for len(m.locals) < int(op.point) {
		m.locals = append(m.locals, uint64(0))
	}
	m.locals[op.point] = m.popFromStack()
}

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
		vm := newVirtualMachine(m.vmCode[m.pointInCode+1:lastPoint], m.contractStorage)
		vm.vmStack = m.vmStack
		vm.debugStack = m.debugStack
		vm.run()
		for i := 0; i < len(vm.vmStack); i++ {
			m.vmStack = append(m.vmStack, vm.vmStack[i])
		}

	} else if op.elsePoint != 0 {
		//do their else statement
		vm := newVirtualMachine(m.vmCode[op.elsePoint:op.endPoint], m.contractStorage)
		vm.vmStack = m.vmStack
		vm.debugStack = m.debugStack
		vm.run()
		for i := 0; i < len(vm.vmStack); i++ {
			m.vmStack = append(m.vmStack, vm.vmStack[i])
		}
	}
	m.pointInCode = uint64(op.endPoint)
}
