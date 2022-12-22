package vm

import (
	"math/big"
)

type ChainDataHandler interface {
	getAddress() []byte
	getBalance([]byte) big.Int //1nite takes up 67 bits(ish), 2 uint64s get pushed to stack. (arbitrary limit. 2^128 is hopefully enough total storage of value)
	getCallerAddress() []byte
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
	//pops address off stack (3 uint64s), pushes balance as 2 uint64s to the stack.
	addressUints := make([]uint64, 3)
	addressUints[2] = m.popFromStack() //needs to be added in reverse order...
	addressUints[1] = m.popFromStack()
	addressUints[0] = m.popFromStack()

	balanceInts := balanceToArray(m.chainHandler.getBalance(uintsArrayToAddress(addressUints)))
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

func (op callerAddr) doOp(m* Machine) error {
	m.pushToStack(m.chainHandler.getAddress())

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
	m.pushToStack(m.blockCtx.Time) // Requires the right encoding
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
}

type valueOp struct {
	gas uint64
}

func (op valueOp) doOp(m *Machine) error {

}


type gasPrice struct {
	gas uint64
}

func (op gasPrice) doOp(m *Machine) error {

}

type codeSize struct {
	gas uint64
}

func (op codeSize) doOp(m *Machine) error {
}

type getCode struct {
	gas uint64
}

func (op getCode) doOp(m *Machine) error {
}

type copyCode struct {
	gas uint64
}

func (op copyCode) doOp(m *Machine) error {
}

type getData struct {
	gas uint64
}

func (op getData) doOp(m *Machine) error {
}

