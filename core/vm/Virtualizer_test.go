package vm

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	apiEndpoint            = "http://0.0.0.0:5001/"
	addTwoFunctionCode     = "0061736d0100000001070160027f7f017f03020100070a010661646454776f00000a09010700200020016a0b000a046e616d650203010000"
	addTwoFunctionBytes, _ = hex.DecodeString(addTwoFunctionCode)
	// addTwoFunctionHash     = hex.EncodeToString(crypto.MD5.New().Sum(addTwoFunctionBytes))
	// .New().Write(addTwoFunctionBytes)
	addTwoFunctionHash = "cee781f77fd0297aae7e71ae0a5d23ba"
	testContract       = ContractData{
		Address: "1",
		Methods: []string{addTwoFunctionHash},
		Storage: []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}
)

func TestBasics(t *testing.T) {
	v := newVirtualizerFromAPI(apiEndpoint, "1", nil)

	stackOut, _, err := v.run(0, []uint64{1, 2})
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(stackOut)
	assert.Equal(t, stackOut, []uint64{3})
}

func TestSpoofedDB(t *testing.T) {
	spoofed := DBSpoofer{make(map[string][]byte)}
	spoofed.addSpoofedCode(addTwoFunctionHash, addTwoFunctionBytes)

	v := newVirtualizerFromData(testContract, &VirtualizerConfig{true, spoofed.GetCode, spoofed.GetCodeBytes})
	stack, _, err := v.run(0, []uint64{5, 6})
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	assert.Equal(t, stack, []uint64{11})

	stack, _, err = v.run(0, []uint64{0xFF, 0x01})
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	assert.Equal(t, stack, []uint64{0x100})
}

func TestUploadingContract(t *testing.T) {
	//this is returning a pass even with the api offline???
	err := uploadContract(apiEndpoint, testContract)
	if err != nil {
		fmt.Println("FAILED")
		fmt.Println(err)
		t.Fail()
	}
}

func TestGettingContract(t *testing.T) {
	cdata, err := getContractData(apiEndpoint, testContract.Address)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	if !contractsEqual(*cdata, testContract) {
		t.Fail()
	}
	// fmt.Println("Contract retrieved is ")
	// fmt.Println(*cdata)

}

func TestUploadingCode(t *testing.T) {
	fmt.Println(addTwoFunctionHash)
	hash, err := uploadMethodString(apiEndpoint, addTwoFunctionCode)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	//assert the hex string format of the hash
	assert.Equal(t, hex.EncodeToString(hash), addTwoFunctionHash)

}
func TestGettingCode(t *testing.T) {
	codeString, err := getMethodCode(apiEndpoint, addTwoFunctionHash)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	assert.Equal(t, hex.EncodeToString(codeString), addTwoFunctionCode)
}
