package vm

import (
	"math/big"
)

type ChainDataHandler interface {
	getAddress() []byte
	getBalance() big.Int //1nite takes up 67 bits(ish), 2 uint64s get pushed to stack. (arbitrary limit. 2^128 is hopefully enough total storage of value)
	getCallerAddress() []byte
	getCallerBalance() big.Int
	getBlockTimestamp() []byte
}
type opAddress struct {
	gas uint64
}

func (op opAddress) doOp(m *Machine) error {
	//addresses are 160 bit, so we need to take the address of the contract, split it into 2 uint64s and a uint32
	// addressBytes := m.chainHandler.getAddress()
	addressBytes := m.chainHandler.getAddress()
	addressInts := addressToInts(addressBytes)
	m.pushToStack(addressInts[0])
	m.pushToStack(addressInts[1])
	m.pushToStack(addressInts[2])

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
	//pushes balance as 3 uint64s to the stack.
	balanceInts := balanceToArray(m.chainHandler.getBalance())
	for i := range balanceInts {
		m.pushToStack(balanceInts[i])
	}
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}
