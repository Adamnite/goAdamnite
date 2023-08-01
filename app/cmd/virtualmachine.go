package cmd

import (
	"fmt"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/common"
)

type VmHandler struct {
	ocdbLink    string
	debuggingDB *VM.SpoofedDBCache
	trueDBCache *VM.DBCache
}

// creates a new vm handler
func NewVMHandler(dblink string) *VmHandler {
	vmh := VmHandler{
		ocdbLink: dblink,
		// localDB: ,
	}
	VM.NewDBCache(dblink)
	VM.NewSpoofedDBCache(nil, nil)

	return &vmh
}

func (vmh *VmHandler) GetCommands() *ishell.Cmd {
	vmFuncs := ishell.Cmd{
		Name: "vm",
		Help: "run all VM options needed for local testing, or submission to the chain",
	}

	vmFuncs.AddCmd(&ishell.Cmd{
		Name: "setLocalState",
		Aliases: []string{
			"set-local-state",
			"sls",
			"set-contract-state",
			"setContractState",
		},
		Help: "setLocalState <contractAddress> <contractStateFile>",
		LongHelp: "" +
			"setLocalState <contractAddress> <contractStateFile>\n" +
			"set the local state of a contract. Useful for debugging contracts by setting the state of the contract to any testable point in time\n" +
			"the changed state should be cleared from the local cache before uploading the changes to the DB",
		Func: vmh.SetLocalStateFromCLI,
	})
	vmFuncs.AddCmd(&ishell.Cmd{
		Name: "getState",
		Help: "getState <contract Address> gets the address of a contract",
		Func: vmh.GetContractFromDBCLI,
	})
	return &vmFuncs
}
func (vmh *VmHandler) SetLocalStateFromCLI(c *ishell.Context) {
	//get the contract address as the first parameter
	//then get the state, and pass that to the SetLocalState with the parameters
	if len(c.Args) > 2 {
		c.Println("Not enough arguments passed")
		return
	}
}
func (vmh *VmHandler) SetLocalState(at common.Address, contract *VM.Contract) error {
	vmh.debuggingDB.DB.AddContract(at.Hex(), contract)
	return nil
}

func (vmh *VmHandler) GetContractFromDBCLI(c *ishell.Context) {
	//the address should be passed.
	if len(c.Args) > 1 {
		c.Println("Not enough arguments passed")
		return
	}
	var contractAddress common.Address
	addressString := strings.ToLower(c.Args[0])
	if addressString[0] == '0' && addressString[1] == 'x' {
		//hex
		contractAddress = common.HexToAddress(addressString)
	} else {
		//string formatted address
		contractAddress = common.StringToAddress(addressString)
	}
	con, err := vmh.GetContractFromDB(contractAddress)
	if err != nil {
		c.Printf("could not get contract @%v, due to the following error\n%v\n", contractAddress.Hex(), err)
		return
	}
	//print the contract information in a cool way.
	c.Println(formatContractAsString(con))
}

func (vmh *VmHandler) GetContractFromDB(at common.Address) (*VM.Contract, error) {
	//get the contract from the DB. Return any errors that occur.
	return vmh.debuggingDB.GetContract(at.Hex())
}
func formatContractAsString(con *VM.Contract) string {
	ans := fmt.Sprintf("\n"+
		"Contract @%v"+
		"Live Hash    %v"+
		"Method Hashes \n\t\t%v"+
		"",
		con.Address.Hex(),
		con.Hash().Hex(),
		strings.Join(con.CodeHashes, "\n\t\t"),
	)
	return ans
}
