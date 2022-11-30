package vm

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	apiEndpoint        = "http://0.0.0.0:5001/"
	addTwoFunctionCode = "0061736d0100000001070160027f7f017f03020100070a010661646454776f00000a09010700200020016a0b000a046e616d650203010000"
	addTwoFunctionHash = "372a47e1d5575acbcff5250366b9abadf73fba0ff1eb3927747ebbf6b7ffe23325c53c0c7281aae31d80703c9d3f75b5e41aaf4cc82985f306b81d882c72995b"
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

func TestUploadingContract(t *testing.T) {
	//this is returning a pass even with the api offline???

	cdata := ContractData{
		Address: "1",
		Methods: []string{addTwoFunctionHash},
		Storage: []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}
	err := uploadContract(apiEndpoint, cdata)
	if err != nil {
		fmt.Println("FAILED")
		fmt.Println(err)
		t.Fail()
	}
}

func TestGettingContract(t *testing.T) {
	cdata, err := getContractData(apiEndpoint, "1")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	// fmt.Println("Contract retrieved is ")
	// fmt.Println(*cdata)
	if cdata.Address != "1" {
		t.Fail()
	}
}

func TestUploadingCode(t *testing.T) {
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
