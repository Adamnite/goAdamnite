package vm

import (
	"fmt"
)

type ChainDataHandler interface {
	getAddress() []byte
	getBalance() []byte
	getCallerAddress() []byte
	getCallerBalance() []byte
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
	balance := m.chainHandler.getBalance()
	fmt.Println("YOU CANT ACTUALLY CHECK BALANCES LIKE THIS YET!!!")
	fmt.Print("but, the balance passed is :")
	fmt.Println(balance)
	if !m.useGas(op.gas) {
		return ErrOutOfGas
	}

	m.pointInCode++
	return nil
}
