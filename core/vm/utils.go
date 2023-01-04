package vm

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/vmihailenco/msgpack/v5"
)

type CodeStored struct {
	CodeParams  []ValueType
	CodeResults []ValueType
	CodeBytes   []byte
}

// API DB Spoofing
type DBSpoofer struct {
	storedFunctions map[string]CodeStored //hash=>functions
}

func newDBSpoofer() DBSpoofer {
	return DBSpoofer{map[string]CodeStored{}}
}

func (spoof *DBSpoofer) GetCode(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
	localCode := spoof.storedFunctions[hex.EncodeToString(hash)]
	ops, blocks := parseBytes(localCode.CodeBytes)

	funcType := FunctionType{
		params:  localCode.CodeParams,
		results: localCode.CodeResults,
		string:  hex.EncodeToString(hash), //so you can lie better.
	}
	return funcType, ops, blocks
}

func (spoof *DBSpoofer) addSpoofedCode(hash string, funcCode CodeStored) {
	spoof.storedFunctions[hash] = funcCode
}
func (spoof *DBSpoofer) addModuleToSpoofedCode(mod Module) (error, [][]byte) {
	hashes := [][]byte{}
	for x := range mod.functionSection {
		code := CodeStored{
			CodeParams:  mod.typeSection[x].params,
			CodeResults: mod.typeSection[x].results,
			CodeBytes:   mod.codeSection[x].body,
		}
		localHash, err := code.hash()

		if err != nil {
			return err, nil
		}
		spoof.addSpoofedCode(hex.EncodeToString(localHash), code)
		hashes = append(hashes, localHash)
	}
	return nil, hashes
}

func (spoof *DBSpoofer) GetCodeBytes(hash string) ([]byte, error) {
	return spoof.storedFunctions[hash].CodeBytes, nil
}

func (spoof *DBSpoofer) getCode2CallName(hash string, inputs []uint64) string {
	ansString := hash //function identifier
	cs := spoof.storedFunctions[hash]
	for i := 0; i < len(cs.CodeParams); i++ {
		ansString += "42" + (hex.EncodeToString(LE.AppendUint64([]byte{}, inputs[i]))[2:])
	} //TODO: write this properly to actually take the param type into account.
	return ansString
}

type BCSpoofer struct {
	contractAddress []byte
	balances        map[string]big.Int
	callerAddress   []byte
	callBlockTime   []byte
}

func newBCSpoofer() BCSpoofer {
	spoofer := BCSpoofer{}
	spoofer.contractAddress = []byte{}
	spoofer.balances = make(map[string]big.Int)
	return spoofer
}

func (s BCSpoofer) setBalanceFromByteAddress(address []byte, balance big.Int) {
	s.balances[hex.EncodeToString(address)] = balance
}
func (s BCSpoofer) setBalance(address string, balance big.Int) {
	s.balances[address] = balance
}

func (s BCSpoofer) getAddress() []byte {
	return s.contractAddress
}
func (s BCSpoofer) getBalance(address []byte) big.Int {
	return s.balances[hex.EncodeToString(address)]
}
func (s BCSpoofer) getCallerAddress() []byte {
	return s.callerAddress
}
func (s BCSpoofer) getBlockTimestamp() []byte {
	return s.callBlockTime
}

//GENERALLY USEFUL

func addressToInts(address []byte) []uint64 {
	//if an empty address is returned, it should take the same amount of space as a full one.
	//converts and address to an array of uint64s, that way it can be pushed to the stack with more ease
	ans := []uint64{0, 0, 0}
	for len(address) <= 24 { //192 bits
		address = append(address, 0)
	}
	ans[0] = LE.Uint64(address[:8]) //yes, this could be a loop, but its more annoying that way.
	address = address[8:]
	ans[1] = LE.Uint64(address[:8])
	address = address[8:]
	ans[2] = LE.Uint64(address[:8])
	return ans
}
func uintsArrayToAddress(input []uint64) []byte {
	ans := []byte{}
	ans = LE.AppendUint64(ans, input[0])
	ans = LE.AppendUint64(ans, input[1])
	ans = LE.AppendUint64(ans, input[2])
	ans = ans[:20]
	return ans
}
func balanceToArray(input big.Int) []uint64 {
	//takes a big int and returns an array of LE formatted uint64s
	b := flipEndian(input.Bytes())

	//b is now in little endian format
	for len(b) < 16 { //add trailing 0s until it is certainly the correct size
		b = append(b, 0)
	}
	leUints := []uint64{
		LE.Uint64(b[0:8]),
		LE.Uint64(b[8:16]),
	}

	return leUints
}
func arrayToBalance(input []uint64) *big.Int {
	//big int uses big endian, so we need to convert back...

	//first convert it to bytes array
	leBytes := []byte{}
	for x := range input {
		leBytes = LE.AppendUint64(leBytes, input[x])
	}
	//then flip and convert to big int!
	ans := big.NewInt(0).SetBytes(flipEndian(leBytes))
	return ans
}

func flipEndian(b []byte) []byte {
	//takes bytes of one endian type, then returns it in the other (flipped)
	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-i-1] = b[len(b)-i-1], b[i]
	}
	return b
}

func contractsEqual(a Contract, b Contract) bool {
	if a.Address != b.Address {
		return false
	}

	for i := range a.Code {
		hashA, err := a.Code[i].hash()
		hashB, errB := b.Code[i].hash()
		if err != nil || errB != nil || !bytes.Equal(hashA, hashB) {
			return false
		}
	}
	for i := range a.Storage {
		if a.Storage[i] != b.Storage[i] {
			return false
		}
	}
	return true
}

//GETTER

func getMethodCode(apiEndpoint string, codeHash string) (*CodeStored, error) {
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

	var code CodeStored
	err = msgpack.Unmarshal(byteResponse, &code)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &code, nil
}

func getContractData(apiEndpoint string, contractAddress string) (*Contract, error) {
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
	var contractData Contract
	err = msgpack.Unmarshal(byteResponse, &contractData)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &contractData, nil
}

//UPLOADER

func uploadMethod(apiEndpoint string, code CodeStored) ([]byte, error) {
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
	byteResponse, err := ioutil.ReadAll(ans.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println("code upload response")
	fmt.Println(hex.EncodeToString(byteResponse))
	fmt.Println(byteResponse)
	fmt.Println(string(byteResponse))
	hashInBytes, err := hex.DecodeString(string(byteResponse))
	if err != nil {
		return nil, err
	}
	return hashInBytes, nil
}

func uploadContract(apiEndpoint string, cdata Contract) error {
	contractApiString := apiEndpoint
	if contractApiString[len(contractApiString)-1:] == "/" {
		contractApiString = contractApiString[:len(contractApiString)-1]
	}

	packedData, err := msgpack.Marshal(&cdata)
	if err != nil {
		fmt.Println(err)
		return err
	}

	re, err := http.NewRequest("PUT", contractApiString+"/contract/"+cdata.Address.String(), bytes.NewReader(packedData))
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

func uploadModuleFunctions(apiEndpoint string, mod Module) ([]CodeStored, [][]byte, error) {
	functionsToUpload := []CodeStored{}
	hashes := [][]byte{}
	for x := range mod.typeSection {
		code := CodeStored{
			CodeParams:  mod.typeSection[x].params,
			CodeResults: mod.typeSection[x].results,
			CodeBytes:   mod.codeSection[x].body,
		}
		functionsToUpload = append(functionsToUpload, code)
		newHash, err := uploadMethod(apiEndpoint, code)
		if err != nil {
			return nil, nil, err
		}
		localHash, err := code.hash()
		if bytes.Equal(newHash, localHash) || err != nil {
			fmt.Println(err)
			return nil, nil, fmt.Errorf("hashes are not equal, or could not hash local copy. ERR: %w, server hash: %v, local hash: %v", err, newHash, localHash)
		}
		hashes = append(hashes, newHash)
	}

	return functionsToUpload, hashes, nil
}

func (code CodeStored) hash() ([]byte, error) {
	packedData, err := msgpack.Marshal(&code)
	if err != nil {
		return nil, err
	}
	hasher := md5.New()
	hasher.Write(packedData)
	return hasher.Sum(nil), nil
}
