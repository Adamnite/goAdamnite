package vm


//This is some pseudocode/POC for what the VM environment could look like. The VM environment will allow for access to core blockchain data, such as state and blocktime.

import (
	"math/big"
	//Import relevant VM and core state modules
	"github.com/adamnite/goadamnite/types"
	"github.com/adamnite/goadamnite/vm"
	//Import others as neccescary
)


//Core environment
type environment struct {
	chain_state *state.chain //Any reference to the state of the core blockchain
	db_state *db.db_state //Any reference to the state of the offchain database
	block *types.Block
	data Data
	typ  vm.Type
}

func newenv(chain_state *state.chain, db_state *db.db_state, block *types.block, data Data) *environment{
	return &environment{
		chain_state: chain_state,
		db_state: db_state,
		block: block,
		data: data,
		typ vm.Type
	}
}

func (self *environment) Origin() common.Address { f, _ := self.data.From(); return f}
func (self *environment) Block_Nonce() *big.Int {return self.block.Number}
func (self *environment) Witness() common.Address { return self.block.witness}
func (self *environment) Timestamp() int64 {return block.timestamp()}
func (self *environment) Value() *big.Int          { return self.data.Value() }
func (self *environment) State() *state.StateDB    { return self.state }

//Transfer, call, and basic execution functions

func (self *environment) Transfer(from, to vm.Account, amount *big.Int) error{
	return vm.Transfer(from, to, amount)
}

func (self *environment) Call(me vm.ContextRef, addr common.Address, data []byte, ate, price, value *big.Int) ([]byte, error) {
	exe := NewExecution(self, &addr, data, ate, price, value)
	return exe.Call(addr, me)
}
func (self *environment) CallCode(me vm.ContextRef, addr common.Address, data []byte, ate, price, value *big.Int) ([]byte, error) {
	maddr := me.Address()
	exe := NewExecution(self, &maddr, data, ate, price, value)
	return exe.Call(addr, me)
}

func (self *environment) Create(me vm.ContextRef, data []byte, ate,  price, value *big.Int) ([]byte, error, vm.ContextRef) {
	exe := NewExecution(self, nil, data, ate, price, value)
	return exe.Create(me)
}
