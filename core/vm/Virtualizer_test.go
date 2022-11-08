package vm

import (
	"fmt"
	"testing"
)

func TestBasics(t *testing.T) {
	v := newVirtualizerFromAPI("http://0.0.0.0:5001/", "1234567890", nil)
	fmt.Println(v)

}

func TestUploadingContractData(t *testing.T) {
	//this is returning a pass even with the api offline???

	cdata := ContractData{
		address: "1",
		methods: []string{"hello"},
		storage: []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}
	err := uploadContract("http://0.0.0.0:5001/", cdata)
	if err != nil {
		fmt.Println("FAILED")
		fmt.Println(err)
		t.Fail()
	}
	// fmt.Println("HELLO? ??")
}

func TestGettingContractData(t *testing.T) {
	cdata, err := getContractData("http://0.0.0.0:5001/", "1")
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println("Contract retrieved is ")
	fmt.Println(cdata)
}

func TestGetMethodCode(t *testing.T) {

}
