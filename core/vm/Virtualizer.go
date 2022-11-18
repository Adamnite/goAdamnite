package vm

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/vmihailenco/msgpack/v5"
)

//the goal of the virtualizer is to make the VM more idiot proof. There should not be a way (without calling the virtualizer.vm)
//to have things go wrong, ideally if all parameters are passed correctly, this should just work.

type VirtualizerInterface interface {
	run(int, []uint64) ([]uint64, []byte, error)
}

type VirtualizerConfig struct {
	saveStateChangeToContractData bool
}

type Virtualizer struct {
	// vm           Machine
	uri          string
	contractData ContractData
	config       VirtualizerConfig
}

// type ContractInterface interface{}
type ContractData struct {
	address string   //the address of the contract
	methods []string //just store all the hashes of the functions it can run as strings
	storage []byte   //all storage inside the contract is held as an array of bytes
}

func (v Virtualizer) run(methodIndex int, locals []uint64) ([]uint64, []byte, error) {
	//the index in the method list that you wish to call
	//returns the stack, storage, and errors that may have occurred.

	//TODO: check that all needed variables are established, and return the stack and storage of the vm
	vmCode, err := getMethodCode(v.uri, v.contractData.methods[methodIndex])
	if err != nil {
		return nil, nil, err
	}
	vm := newVirtualMachine(
		vmCode,
		v.contractData.storage,
		nil, 
		1000) //config should be created/have defaults. @TODO update this with right Gas value
	vm.locals = locals
	vm.run()

	if v.config.saveStateChangeToContractData {
		v.contractData.storage = vm.contractStorage
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
	}
	return v
}
func (v *Virtualizer) setVirtualizerConfig(config *VirtualizerConfig) {
	if config == nil { //set default config data
		v.config = VirtualizerConfig{
			true}
	} else {
		v.config = *config
	}
}
func getContractData(apiEndpoint string, contractAddress string) (*ContractData, error) {
	contractApiString := apiEndpoint
	if contractApiString[len(contractApiString)-1:] == "/" {
		contractApiString = contractApiString[:len(contractApiString)-1]
	}
	re, err := http.Get(contractApiString + "/contract/" + contractAddress)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	byteResponse, err := ioutil.ReadAll(re.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//hopefully you know if things went wrong by here!
	contractData := ContractData{}
	err = msgpack.Unmarshal(byteResponse, &contractData)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &contractData, nil
}

func getMethodCode(apiEndpoint string, codeHash string) ([]byte, error) {
	ApiString := apiEndpoint
	if ApiString[len(ApiString)-1:] == "/" {
		ApiString = ApiString[:len(ApiString)-1]
	}
	re, err := http.Get(apiEndpoint + "/code/" + codeHash)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	byteResponse, err := ioutil.ReadAll(re.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return byteResponse, nil
}
func uploadContract(apiEndpoint string, cdata ContractData) error {
	contractApiString := apiEndpoint
	if contractApiString[len(contractApiString)-1:] == "/" {
		contractApiString = contractApiString[:len(contractApiString)-1]
	}

	packedData, err := msgpack.Marshal(&cdata)
	if err != nil {
		fmt.Println("AHHHH")
		fmt.Println(err)
		return err
	}
	fmt.Println(cdata)
	fmt.Println()
	println("Packed data")
	fmt.Println(packedData)

	var testCdata ContractData
	err = msgpack.Unmarshal(packedData, &testCdata)
	fmt.Println(err)
	fmt.Println(testCdata)

	fmt.Println("bytes as reader")
	fmt.Println(bytes.NewReader(packedData))
	re, err := http.NewRequest("PUT", contractApiString+"/contract/"+cdata.address, bytes.NewReader(packedData))
	if err != nil {
		fmt.Println(err)
		return err
	}
	// re.
	// fmt.Println(re.)
	ans, err := http.DefaultClient.Do(re)
	if err != nil {
		fmt.Println(err)
		fmt.Println(ans)
		return err
	}
	return nil

}
