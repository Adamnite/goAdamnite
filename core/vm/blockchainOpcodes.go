package vm

import (
	"github.com/adamnite/go-adamnite/common"
)

type opAddress struct {
	gas uint64
}

func (op opAddress) doOp(m *Machine) error {
	//addresses are 224 bit, so we need to take the address of the contract, split it into 3 uint64s
	// addressBytes := m.chainHandler.getAddress()
	addressBytes := m.contract.Address.Bytes()
	addressInts := addressToInts(addressBytes)
	for i := 0; i < len(addressInts); i++ {
		m.pushToStack(addressInts[i])
	}

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type balance struct {
	gas uint64
}

func (op balance) doOp(m *Machine) error {
	//pops address off stack (4 uint64s), pushes balance as 2 uint64s to the stack.
	addressUints := make([]uint64, 4)
	addressUints[3] = m.popFromStack() //needs to be added in reverse order...
	addressUints[2] = m.popFromStack()
	addressUints[1] = m.popFromStack()
	addressUints[0] = m.popFromStack()

	addr := uintsArrayToAddress(addressUints)
	value := m.Statedb.GetBalance(common.BytesToAddress(addr))
	balanceInts := balanceToArray(*value)
	for i := range balanceInts {
		m.pushToStack(balanceInts[i])
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type callerAddr struct {
	gas uint64
}

func (op callerAddr) doOp(m *Machine) error {
	addressBytes := m.contract.CallerAddress.Bytes()
	addressInts := addressToInts(addressBytes)
	for i := 0; i < len(addressInts); i++ {
		m.pushToStack(addressInts[i])
	}

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type blocktimestamp struct {
	gas uint64
}

func (op blocktimestamp) doOp(m *Machine) error {
	ts := EncodeUint64(uint64(m.BlockCtx.Time.Int64()))
	m.pushToStack(ts)
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type dataSize struct {
	gas uint64
}

func (op dataSize) doOp(m *Machine) error {
	size := uint64(len(m.contract.Input))

	m.pushToStack(EncodeUint64(size))
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}

type valueOp struct {
	gas uint64
}

func (op valueOp) doOp(m *Machine) error {
	v := m.contract.Value

	m.pushToStack(balanceToArray(*v))
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type gasPrice struct {
	gas uint64
}

func (op gasPrice) doOp(m *Machine) error {
	v := m.txCtx.GasPrice

	m.pushToStack(balanceToArray(*v))

	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type codeSize struct {
	gas uint64
}

func (op codeSize) doOp(m *Machine) error {
	return nil
}

type getCode struct {
	gas uint64
}

func (op getCode) doOp(m *Machine) error {
	return nil
}

type copyCode struct {
	gas uint64
}

func (op copyCode) doOp(m *Machine) error {
	return nil
}

type getData struct {
	gas uint64
}

func (op getData) doOp(m *Machine) error {
	data := m.contract.Input
	m.pushToStack(data)
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}
