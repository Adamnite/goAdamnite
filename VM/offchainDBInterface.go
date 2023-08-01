package VM

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrContractNotStored = fmt.Errorf("no contract saved at that point")
)

type APIcodeGetter struct {
	apiEndpointString string
}

func NewAPICodeGetter(apiString string) APIcodeGetter {
	return APIcodeGetter{apiEndpointString: apiString}
}

func (c APIcodeGetter) GetCode(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
	locCopy, err := GetMethodCode(c.apiEndpointString, hex.EncodeToString(hash))
	if err != nil {
		panic(err)
	}
	ops, blocks := parseBytes(locCopy.CodeBytes)
	funcType := FunctionType{
		params:  locCopy.CodeParams,
		results: locCopy.CodeResults,
		string:  hex.EncodeToString(hash), //so you can lie better.
	}
	return funcType, ops, blocks
	// func GetMethodCode(apiEndpoint string, codeHash string) (*CodeStored, error) {
}

//UPLOADER

func UploadMethod(apiEndpoint string, code CodeStored) ([]byte, error) {
	//takes the code as an array of bytes and returns the hash, and any errors
	contractApiString := apiEndpoint
	if contractApiString[len(contractApiString)-1:] == "/" {
		contractApiString = contractApiString[:len(contractApiString)-1]
	}
	packedData, err := msgpack.Marshal(&code)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	re, err := http.NewRequest("PUT", contractApiString+"/uploadCode", bytes.NewReader(packedData))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	ans, err := http.DefaultClient.Do(re)
	if err != nil {
		return nil, err
	}
	netErr := NewErrWithNetwork(ans.StatusCode)
	if !netErr.Is(200) {
		return nil, netErr
	}

	byteResponse, err := io.ReadAll(ans.Body)
	if err != nil {
		return nil, err
	}
	hashInBytes, err := hex.DecodeString(string(byteResponse))
	if err != nil {
		return nil, err
	}
	return hashInBytes, nil
}

func UploadContract(apiEndpoint string, con Contract) error {
	contractApiString := apiEndpoint
	if contractApiString[len(contractApiString)-1:] == "/" {
		contractApiString = contractApiString[:len(contractApiString)-1]
	}

	packedData, err := ContractToMSGPackBytes(con)
	if err != nil {
		fmt.Println(err)
		return err
	}

	re, err := http.NewRequest("PUT", contractApiString+"/contract/"+con.Address.Hex(), bytes.NewReader(packedData))
	if err != nil {
		fmt.Println(err)
		return err
	}

	ans, err := http.DefaultClient.Do(re)
	if err != nil {
		return err
	}
	netErr := NewErrWithNetwork(ans.StatusCode)
	if !netErr.Is(200) {
		return netErr
	}
	return nil

}

func UploadModuleFunctions(apiEndpoint string, mod Module) ([]CodeStored, [][]byte, error) {
	functionsToUpload := []CodeStored{}
	hashes := [][]byte{}
	for x := range mod.typeSection {
		code := CodeStored{
			CodeParams:  mod.typeSection[x].params,
			CodeResults: mod.typeSection[x].results,
			CodeBytes:   mod.codeSection[x].body,
		}
		functionsToUpload = append(functionsToUpload, code)
		newHash, err := UploadMethod(apiEndpoint, code)
		if err != nil {
			return nil, nil, err
		}
		localHash, err := code.Hash()

		if !bytes.Equal(newHash, localHash) || err != nil {
			return nil, nil, fmt.Errorf("hashes are not equal, or could not hash local copy. ERR: %w, server hash: %v, local hash: %v", err, newHash, localHash)
		}
		hashes = append(hashes, newHash)
	}

	return functionsToUpload, hashes, nil
}
func (m Machine) UploadMachinesContract(apiEndpoint string) error {
	return UploadContract(apiEndpoint, m.contract)
}

//GETTER

func GetMethodCode(apiEndpoint string, codeHash string) (*CodeStored, error) {
	ApiString := apiEndpoint
	if ApiString[len(ApiString)-1:] == "/" {
		ApiString = ApiString[:len(ApiString)-1]
	}
	re, err := http.Get(ApiString + "/code/" + codeHash)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	byteResponse, err := io.ReadAll(re.Body)
	if err != nil {
		return nil, err
	}

	var code CodeStored
	err = msgpack.Unmarshal(byteResponse, &code)
	if err != nil {
		return nil, err
	}

	return &code, nil
}

func GetContractData(apiEndpoint string, contractAddress string) (*Contract, error) {
	contractApiString := apiEndpoint
	if contractApiString[len(contractApiString)-1:] == "/" {
		contractApiString = contractApiString[:len(contractApiString)-1]
	}
	re, err := http.Get(contractApiString + "/contract/" + contractAddress)
	if err != nil {
		return nil, err
	}

	byteResponse, err := io.ReadAll(re.Body)
	if err != nil {
		return nil, err
	}
	if string(byteResponse) == "contract not stored" {
		return nil, ErrContractNotStored
	}
	//hopefully you know if things went wrong by here!
	var conData ContractData
	err = msgpack.Unmarshal(byteResponse, &conData)
	if err != nil {
		return nil, err
	}
	foo := contractDataToContract(conData)
	return foo, nil
}
