package VM

import (
	"encoding/hex"
	"errors"
	"reflect"
)

type Block struct {
	index uint32 // The index of its controlblock inside the controlBlockStack
	gas   uint64
}

func (op Block) doOp(m *Machine) error {
	// Add some stack validation here
	stackLength := len(m.vmStack)
	currentFrame := m.callStack[m.currentFrame]

	m.pointInCode++ // First skip this Block byte
	control := m.controlBlockStack[op.index]

	for m.pointInCode < control.endAt {

		if m.stopSignal {
			return nil
		}

		op := m.vmCode[m.pointInCode]
		op.doOp(m)

		if reflect.TypeOf(op) == reflect.TypeOf(Call{}) {
			m.stopSignal = true
		}
	}

	currentFrame.Ip = m.pointInCode
	finalStackLength := len(m.vmStack)

	if finalStackLength < stackLength {
		return ErrStackConsistency
	}

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type Br struct {
	index uint32
	gas   uint64
}

func (op Br) doOp(m *Machine) error {

	// An important note about the labels is that the innermost one has the index 0 and the outtermost one has the index N
	// https://webassembly.github.io/spec/core/bikeshed/index.html#control-instructions%E2%91%A0
	if len(m.controlBlockStack) < int(op.index) {
		panic("Index where to branch out of range")
	}

	branch := m.controlBlockStack[len(m.controlBlockStack)-int(op.index)-1]
	currentFrame := m.callStack[m.currentFrame]

	if branch.op == Op_block || branch.op == Op_if {
		// This means a break statement
		m.pointInCode = branch.endAt
		currentFrame.Ip = branch.endAt
	} else if branch.op == Op_loop {
		// This means a continue statement
		m.pointInCode = branch.startAt + 1 // +1 To skip the block byte
		currentFrame.Ip = branch.startAt
	} else {
		return ErrInvalidBr
	}

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type BrIf struct {
	index uint32
	gas   uint64
}

func (op BrIf) doOp(m *Machine) error {
	if len(m.controlBlockStack) < int(op.index) {
		return ErrInvalidBr
	}

	condition := uint32(m.popFromStack())

	if condition != 0 {
		Br{op.index, GasQuickStep}.doOp(m)
	} else {
		NoOp{}.doOp(m)
	}

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type If struct {
	index uint32 // The index of its controlblock inside the controlBlockStack
	gas   uint64
}

func (op If) doOp(m *Machine) error {
	// @TODO(sdmg15) Check if the top of the stack is the same type as signature in If/Else/End
	currentFrame := m.callStack[m.currentFrame]
	condition := uint32(m.popFromStack())

	stackLen := len(m.vmStack)

	controlBlock := m.controlBlockStack[int(op.index)]

	if controlBlock.op != Op_if {
		return ErrIfTopElementOfStack
	}

	if condition != 0 {
		end := controlBlock.endAt

		if controlBlock.elseAt != 0 {
			end = controlBlock.elseAt - 1 //
		}

		m.pointInCode++
		for m.pointInCode <= end {

			if m.stopSignal {
				return nil
			}

			op := m.vmCode[m.pointInCode]
			op.doOp(m)
			if reflect.TypeOf(op) == reflect.TypeOf(Call{}) {
				m.stopSignal = true
			}
		}

		if controlBlock.elseAt != 0 {
			m.pointInCode = controlBlock.endAt
			currentFrame.Ip = controlBlock.endAt
		}

	} else if controlBlock.elseAt != 0 {
		m.pointInCode = controlBlock.elseAt + 1 // + 1 to skip the block byte
		currentFrame.Ip = controlBlock.elseAt
	} else {
		m.pointInCode = controlBlock.endAt
		currentFrame.Ip = controlBlock.endAt
	}

	if len(m.vmStack) < stackLen {
		return ErrStackConsistency
	}

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	return nil

}

type Else struct {
	index uint32 // The index of its controlblock inside the controlBlockStack should have the same value as the If one
	gas   uint64
}

func (op Else) doOp(m *Machine) error {

	stackLen := len(m.vmStack)
	controlBlock := m.controlBlockStack[len(m.controlBlockStack)-1]

	for m.pointInCode != controlBlock.endAt {

		if m.stopSignal {
			return nil
		}

		op := m.vmCode[m.pointInCode]
		op.doOp(m)

		if reflect.TypeOf(op) == reflect.TypeOf(Call{}) {
			m.stopSignal = true
		}
	}

	if len(m.vmStack) < stackLen {
		return ErrStackConsistency
	}

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type Loop struct {
	index uint32
	gas   uint64
}

func (op Loop) doOp(m *Machine) error {
	stackLength := len(m.vmStack)
	currentFrame := m.callStack[m.currentFrame]
	m.pointInCode++ // First skip this Loop byte
	controlBlock := m.controlBlockStack[op.index]

	// Once the pointInCode becomes bigger than the endAt then it means we branched to a block
	for m.pointInCode < controlBlock.endAt {

		if m.stopSignal {
			return nil
		}

		op := m.vmCode[m.pointInCode]
		op.doOp(m)

		if reflect.TypeOf(op) == reflect.TypeOf(Call{}) {
			m.stopSignal = true
		}
	}

	currentFrame.Ip = controlBlock.endAt
	finalStackLength := len(m.vmStack)

	if finalStackLength < stackLength {
		return ErrStackConsistency
	}

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type NoOp struct{}

func (op NoOp) doOp(m *Machine) error {
	m.pointInCode++
	return nil
}

type UnReachable struct{}

func (op UnReachable) doOp(m *Machine) error {
	m.pointInCode++
	return nil
}

type End struct {
	index uint32
}

func (op End) doOp(m *Machine) error {
	m.pointInCode++
	return nil
}

type Return struct {
	gas uint64
}

func (op Return) doOp(m *Machine) error {
	// branch := m.controlBlockStack[0]
	currentFrame := m.callStack[m.currentFrame]

	if len(m.vmStack) > 0 {
		res := m.popFromStack()

		for len(m.vmStack) != 0 {
			m.popFromStack()
			m.pointInCode++
		}
		m.pushToStack(uint64(res))
	} else {
		m.pointInCode += uint64(len(m.vmCode) - 2) // -1 for range and -1 for staying at End{} of function
	}

	currentFrame.Ip = m.pointInCode
	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type Call struct {
	funcIndex uint32
	gas       uint64
}

func (op Call) doOp(m *Machine) error {

	if int(op.funcIndex) >= len(m.contract.CodeHashes) {
		return errors.New("invalid function index")
	}

	// Pop the required params from stack
	hexEncodingOfHash, _ := hex.DecodeString((m.contract.CodeHashes[op.funcIndex]))

	//TODO: have this save the code grabbed, if its used multiple times, we shouldn't need to fetch it multiple times.
	lFuncType, lOps, lControlBlocks := m.config.CodeGetter(hexEncodingOfHash)

	params := lFuncType.params
	poppedParams := []uint64{}
	for i := len(params); i != 0; i-- {
		poppedParams = append(poppedParams, m.popFromStack())
	}

	m.callStack[m.currentFrame].Continuation = int64(m.pointInCode) + 1 // When this frame will finish it will load this pointInCode back?
	// Activate the new frame
	frame := new(Frame)
	frame.Code = lOps
	frame.CtrlStack = lControlBlocks
	frame.Locals = poppedParams
	frame.Ip = 0

	m.pointInCode = 0
	m.vmCode = frame.Code
	m.locals = frame.Locals

	m.callStack = append(m.callStack, frame)
	m.currentFrame++

	if !m.useAte(op.gas) {
		return ErrOutOfGas
	}
	return nil
}

type CallIndirect struct{}

func (op CallIndirect) doOp(m *Machine) error {
	return nil
}
