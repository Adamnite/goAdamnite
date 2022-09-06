package vm


type OperationCommon interface {
	doOp(vm *Machine)
}

type Operation struct {

}


func (m *Machine) doAdd() {
	m.pushToStack(m.popFromStack() + m.popFromStack())
}