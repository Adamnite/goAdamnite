package core

import (
	"errors"
	"math/big"

	virtMach "github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/databaseDeprecated/statedb"
	"github.com/adamnite/go-adamnite/common"

	log "github.com/sirupsen/logrus"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for operation fee")
)

type StateTransition struct {
	msg        Message
	ate        uint64
	atePrice   *big.Int
	initialAte uint64
	value      *big.Int
	data       []byte
	state      statedb.StateDB
	vm         *virtMach.Machine
}

// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	//FromFrontier() (common.Address, error)
	To() *common.Address

	AtePrice() *big.Int
	Ate() uint64
	Value() *big.Int

	Nonce() uint64
	CheckNonce() bool
	Data() []byte
}

func NewStateTransition(vmVar *virtMach.Machine, msg Message) *StateTransition {
	return &StateTransition{
		vm:       vmVar,
		msg:      msg,
		atePrice: msg.AtePrice(),
		value:    msg.Value(),
		data:     msg.Data(),

		state: *vmVar.Statedb,
	}
}

func ApplyMessage(vm *virtMach.Machine, msg Message) ([]byte, uint64, bool, error) {
	return NewStateTransition(vm, msg).TransitionDb()
}

// to returns the recipient of the message.
func (st *StateTransition) to() common.Address {
	if st.msg == nil || st.msg.To() == nil /* contract creation */ {
		return common.Address{}
	}
	return *st.msg.To()
}

func (st *StateTransition) useGas(amount uint64) error {
	if st.ate < amount {
		return virtMach.ErrOutOfGas
	}
	st.ate -= amount

	return nil
}

func (st *StateTransition) buyAte() error {
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Ate()), st.atePrice)
	if st.state.GetBalance(st.msg.From()).Cmp(mgval) < 0 {
		return errInsufficientBalanceForGas
	}
	st.ate += st.msg.Ate()

	st.initialAte = st.msg.Ate()
	st.state.SubBalance(st.msg.From(), mgval)
	return nil
}

func (st *StateTransition) TransitionDb() (ret []byte, usedAte uint64, failed bool, err error) {

	msg := st.msg
	sender := msg.From()
	contractCreation := msg.To() == nil

	if err = st.useGas(usedAte); err != nil {
		return nil, 0, false, err
	}

	var (
		vm = st.vm
		// vm errors do not effect consensus and are therefor
		// not assigned to err, except for insufficient balance
		// error.
		vmerr error
	)
	if contractCreation {
		_, st.ate, vmerr = vm.Create(sender, st.data, st.ate, st.value)
	} else {

		// Increment the nonce for the next transaction
		st.state.SetNonce(msg.From(), st.state.GetNonce(sender)+1)
		ret, st.ate, vmerr = vm.Call(
			sender,   //caller address
			st.to(),  //contract address
			st.data,  //function hash followed by function params
			st.ate,   //gas
			st.value) //amount sent to contract
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == virtMach.ErrInsufficientBalance {
			return nil, 0, false, vmerr
		}
	}

	st.refundAte()

	st.state.AddBalance(st.vm.BlockCtx.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.atePrice))

	return ret, st.gasUsed(), vmerr != nil, err
}

func (st *StateTransition) refundAte() {
	// Apply refund counter, capped to half of the used gas.

	refund := st.gasUsed() / 2
	// if refund > st.state.GetRefund() {
	// 	refund = st.state.GetRefund()
	// }
	st.ate += refund

	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.ate), st.atePrice)
	st.state.AddBalance(st.msg.From(), remaining)

}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialAte - st.ate
}
