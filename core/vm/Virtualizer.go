package vm

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	// "golang.org/x/exp/slices"//uncomment if you want to see the imports break...
)

//the goal of the virtualizer is to make the VM more idiot proof. There should not be a way (without calling the virtualizer.vm)
//to have things go wrong, ideally if all parameters are passed correctly, this should just work.

type VirtualizerInterface interface {
	run(int, []uint64) ([]uint64, []byte, error)
}

type VirtualizerConfig struct {
	saveStateChangeToContractData bool
	codeGetter                    GetCode
	codeBytesGetter               GetCodeBytes
	dbLink                        statedb.StateDB
}

type GetCodeBytes func(hash string) ([]byte, error)

type Virtualizer struct {
	// vm           Machine
	uri           string
	contractData  ContractData
	config        VirtualizerConfig
	callerAddress []byte
}

// type ContractInterface interface{}
type ContractData struct {
	Address string   //the Address of the contract
	Methods []string //just store all the hashes of the functions it can run as strings
	Storage []uint64 //all storage inside the contract is held as an array of bytes
}

func (v Virtualizer) run(methodIndex int, locals []uint64, callerAddress []byte) ([]uint64, []uint64, error) {
	//the index in the method list that you wish to call
	//returns the stack, storage, and errors that may have occurred.

	//TODO: check that all needed variables are established, and return the stack and storage of the vm
	v.callerAddress = callerAddress
	vmCode, err := v.config.codeBytesGetter(v.contractData.Methods[methodIndex])
	if err != nil {
		return nil, nil, err
	}
	vm := newVirtualMachine(
		vmCode,
		v.contractData.Storage,
		nil,
		1000) //config should be created/have defaults. @TODO update this with right Gas value
	vm.locals = locals
	vm.callStack[0].Locals = vm.locals
	if vm.chainHandler != nil {
		vm.chainHandler = v
	}
	vm.run()

	if v.config.saveStateChangeToContractData {
		v.contractData.Storage = vm.contractStorage
	}

	return vm.vmStack, vm.contractStorage, nil
}

func newVirtualizerFromAPI(uri string, contractAddress string, config *VirtualizerConfig) Virtualizer {
	//TODO: clean the URI string to make sure it is a direct API call
	v := Virtualizer{
		uri: uri}
	v.setVirtualizerConfig(config)
	cdata, err := getContractData(uri, contractAddress)
	if err != nil {
		fmt.Println("THERE WAS AN ERROR!!!")
		fmt.Println(err)
	}
	v.contractData = *cdata

	return v
}

func newVirtualizerFromData(cdata ContractData, config *VirtualizerConfig) Virtualizer {
	v := Virtualizer{
		contractData: cdata,
		config:       *config,
	}
	return v
}

func (v *Virtualizer) setVirtualizerConfig(config *VirtualizerConfig) {
	if config == nil { //set default config data
		v.config = VirtualizerConfig{
			saveStateChangeToContractData: true,
			codeGetter:                    v.GetCode,
			codeBytesGetter:               v.GetCodeBytes,
		}
	} else {
		v.config = *config
	}
}

func (v *Virtualizer) GetCode(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
	hexHash := hex.EncodeToString(hash)
	//the default code getter, uses the DB to get all functions.
	code, err := getMethodCode(v.uri, hexHash)
	if err != nil {
		panic(err)
	}
	ops, blocks := parseBytes(code.CodeBytes)

	funcType := FunctionType{
		params:  code.CodeParams,
		results: code.CodeResults,
		string:  hexHash,
	}
	// if slices.Contains(v.contractData.Methods, hexHash) {//uncomment if you want to see the imports break...
	// 	panic(fmt.Errorf("attempt to call hash unaccepted by contract with hash:%v", hexHash))
	// }
	inHash := false
	for i := 0; i < len(v.contractData.Methods) || inHash; i++ {
		inHash = hexHash == v.contractData.Methods[i]
	}
	if !inHash {
		panic(fmt.Errorf("attempt to call hash unaccepted by contract with hash:%v", hexHash))
	}

	return funcType, ops, blocks
}

func (v *Virtualizer) GetCodeBytes(hash string) ([]byte, error) {
	code, err := getMethodCode(v.uri, hash)
	if err != nil {
		return nil, err
	}
	return code.CodeBytes, nil
}

func (v Virtualizer) getBalance(address []byte) big.Int {
	return *v.config.dbLink.GetBalance(common.BytesToAddress(address))
}

func (v Virtualizer) getAddress() []byte {
	ans, _ := hex.DecodeString(v.contractData.Address)
	return ans

}
func (v Virtualizer) getCallerAddress() []byte {
	return v.callerAddress
}
func (v Virtualizer) getBlockTimestamp() []byte {
	//TODO: actually get the time stamp
	return []byte{}
}
