package vm

import (
	"bytes"
	"encoding/hex"
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
	Address string   //the Address of the contract
	Methods []string //just store all the hashes of the functions it can run as strings
	Storage []uint64 //all storage inside the contract is held as an array of bytes
}

func (v Virtualizer) run(methodIndex int, locals []uint64) ([]uint64, []uint64, error) {
	//the index in the method list that you wish to call
	//returns the stack, storage, and errors that may have occurred.

	//TODO: check that all needed variables are established, and return the stack and storage of the vm
	vmCode, err := getMethodCode(v.uri, v.contractData.Methods[methodIndex])
	if err != nil {
		return nil, nil, err
	}
	vm := newVirtualMachine(
		vmCode,
		v.contractData.Storage,
		nil,
		1000) //config should be created/have defaults. @TODO update this with right Gas value
	vm.locals = locals
	vm.vmCode, vm.controlBlockStack = parseBytes(*&decode(vmCode).codeSection[0].body)
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
	var contractData ContractData
	err = msgpack.Unmarshal(byteResponse, &contractData)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &contractData, nil
}
func uploadMethodString(apiEndpoint string, code string) ([]byte, error) {
	//takes the string format of the code and returns the hash, and any errors
	byteFormat, err := hex.DecodeString(code)
	if err != nil {
		return nil, err
	}
	return uploadMethod(apiEndpoint, byteFormat)
}
func uploadMethod(apiEndpoint string, code []byte) ([]byte, error) {
	//takes the code as an array of bytes and returns the hash, and any errors
	contractApiString := apiEndpoint
	if contractApiString[len(contractApiString)-1:] == "/" {
		contractApiString = contractApiString[:len(contractApiString)-1]
	}

	re, err := http.NewRequest("PUT", contractApiString+"/uploadCode", bytes.NewReader(code))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	ans, err := http.DefaultClient.Do(re)
	if err != nil {
		return nil, err
	}
	byteResponse, err := ioutil.ReadAll(ans.Body)
	if err != nil {
		return nil, err
	}
	hashInBytes, err := hex.DecodeString(string(byteResponse))
	if err != nil {
		return nil, err
	}
	return hashInBytes, nil
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

	re, err := http.NewRequest("PUT", contractApiString+"/contract/"+cdata.Address, bytes.NewReader(packedData))
	if err != nil {
		fmt.Println(err)
		return err
	}

	ans, err := http.DefaultClient.Do(re)
	if err != nil {
		return err
	}
	if ans.StatusCode != 200 {
		return fmt.Errorf("Host rejected the upload process with reason " + ans.Status)
	}
	return nil

}
