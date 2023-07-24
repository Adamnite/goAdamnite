package VM

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/utils/bytes"
	"github.com/stretchr/testify/assert"
)

var (
	apiEndpoint            = "http://127.0.0.1:5000/"
	addTwoFunctionCode     = "0061736d0100000001070160027f7f017f03020100070a010661646454776f00000a09010700200020016a0b000a046e616d650203010000"
	addTwoFunctionBytes, _ = hex.DecodeString(addTwoFunctionCode)
	addTwoCodeStored       = CodeStored{[]ValueType{Op_i64, Op_i64}, []ValueType{Op_i64}, addTwoFunctionBytes}
	// addTwoFunctionHash     = hex.EncodeToString(crypto.MD5.New().Sum(addTwoFunctionBytes))
	addTwoFunctionHash, _ = addTwoCodeStored.Hash()
	testContract          = newContract(common.BytesToAddress([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}), big.NewInt(0), nil, 10000)
	// testContract          = Contract{
	// 	Address: "1",
	// 	Methods: []string{hex.EncodeToString(addTwoFunctionHash)},
	// 	Storage: []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	// }
	db            = rawdb.NewMemoryDB()
	state, _      = statedb.New(bytes.Hash{}, statedb.NewDatabase(db))
	callerAddress = []byte{0, 1, 2, 3, 4, 5}
)

func TestUploadingContract(t *testing.T) {
	testContract = newContract(common.BytesToAddress([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}), big.NewInt(0), nil, 10000)
	err := UploadContract(apiEndpoint, *testContract)
	fmt.Println(testContract.Address.Hex())
	fmt.Println(testContract)
	if err != nil {
		if err.Error()[len(err.Error())-27:] == "connect: connection refused" {
			fmt.Println("Local Server Appears to be offline.")
			t.SkipNow()
		} else {
			fmt.Println(err)
			t.Fail()
		}
	}
}

func TestGettingContract(t *testing.T) {
	cdata, err := GetContractData(apiEndpoint, testContract.Address.Hex())
	if err != nil {
		if err.Error()[len(err.Error())-27:] == "connect: connection refused" {
			fmt.Println("Local Server Appears to be offline.")
			t.SkipNow()
		} else {
			fmt.Println(err)
			t.Fail()
		}
	}
	if !contractsEqual(*cdata, *testContract) {
		t.Fail()
	}
	// fmt.Println("Contract retrieved is ")
	// fmt.Println(*cdata)

}

func TestUploadingCode(t *testing.T) {
	fmt.Println(addTwoFunctionHash)

	hash, err := UploadMethod(apiEndpoint, addTwoCodeStored)
	if err != nil {
		if err.Error()[len(err.Error())-27:] == "connect: connection refused" {
			fmt.Println("Local Server Appears to be offline.")
			t.SkipNow()
		} else {
			fmt.Println(err)
			t.Fail()
		}

	}
	//assert the hex string format of the hash
	assert.Equal(t, hex.EncodeToString(addTwoFunctionHash), hex.EncodeToString(hash))

}
func TestGettingCode(t *testing.T) {
	codeString, err := GetMethodCode(apiEndpoint, hex.EncodeToString(addTwoFunctionHash))
	if err != nil {
		if err.Error()[len(err.Error())-27:] == "connect: connection refused" {
			fmt.Println("Local Server Appears to be offline.")
			t.SkipNow()
		} else {
			fmt.Println(err)
			t.Fail()
		}
	}
	assert.Equal(t, hex.EncodeToString(codeString.CodeBytes), addTwoFunctionCode)
}
