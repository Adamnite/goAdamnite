package VM

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/adamnite/go-adamnite/common"
	"github.com/vmihailenco/msgpack/v5"
)

// Contract represents an adm contract in the state database. It contains
// the contract methods, calling arguments.
type ContractData struct {
	Address string   //the Address of the contract
	Methods []string //just store all the hashes of the functions it can run as strings
	Storage []uint64 //all storage inside the contract is held as an array of bytes
}
type CodeStored struct {
	CodeParams  []ValueType
	CodeResults []ValueType
	CodeBytes   []byte
}

type DBInterfaceItem interface {
	GetCode([]byte) (FunctionType, []OperationCommon, []ControlBlock)
	GetContract(string) (*Contract, error)
}
type LocalDBTester struct {
	codeGetter     func([]byte) (FunctionType, []OperationCommon, []ControlBlock)
	contractGetter func(string) (*Contract, error)
}

func NewLocalDBTester(
	codeGetter func([]byte) (FunctionType, []OperationCommon, []ControlBlock),
	contractGetter func(string) (*Contract, error),
) *LocalDBTester {
	return &LocalDBTester{
		codeGetter:     codeGetter,
		contractGetter: contractGetter,
	}
}
func (ldbt *LocalDBTester) GetCode(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
	return ldbt.codeGetter(hash)
}
func (ldbt *LocalDBTester) GetContract(address string) (*Contract, error) {
	return ldbt.contractGetter(address)
}

type DBCacheAble interface {
	GetCode(hash []byte) (FunctionType, []OperationCommon, []ControlBlock)
	PreCacheCode(hash []byte) error
	GetContract(address string) (*Contract, error)
	PreCacheContract(address string) error
}

// a thread safe spoofing of a DBCache
type SpoofedDBCache struct {
	lock          sync.Mutex
	DB            *DBSpoofer
	cacheCode     func(hash []byte)
	cacheContract func(address string)
}

func NewSpoofedDBCache(cacheCode func([]byte), cacheContract func(string)) *SpoofedDBCache {
	return &SpoofedDBCache{
		DB:            NewDBSpoofer(),
		cacheCode:     cacheCode,
		cacheContract: cacheContract,
	}
}
func (sdbc *SpoofedDBCache) GetCode(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
	sdbc.lock.Lock()
	defer sdbc.lock.Unlock()
	return sdbc.DB.GetCode(hash)
}
func (sdbc *SpoofedDBCache) PreCacheCode(hash []byte) error {
	sdbc.lock.Lock()
	defer sdbc.lock.Unlock()
	if sdbc.cacheCode != nil {
		sdbc.cacheCode(hash)
	}
	return nil
}
func (sdbc *SpoofedDBCache) GetContract(address string) (*Contract, error) {
	sdbc.lock.Lock()
	defer sdbc.lock.Unlock()
	return sdbc.DB.GetContract(address)
}

func (sdbc *SpoofedDBCache) PreCacheContract(address string) error {
	sdbc.lock.Lock()
	defer sdbc.lock.Unlock()
	if sdbc.cacheContract != nil {
		sdbc.cacheContract(address)
	}
	return nil
}

type DBCache struct {
	// cache acts the same as the DBSpoofer, except it is thread safe, and get's its own code
	funcLock  sync.Mutex
	functions map[string]CodeStored //hash=>functions
	conLock   sync.Mutex
	contracts map[string]*Contract
	api       string
}

func NewDBCache(apiEndpoint string) DBCacheAble {
	return &DBCache{
		functions: map[string]CodeStored{},
		contracts: map[string]*Contract{},
		api:       apiEndpoint,
	}
}

func (cache *DBCache) GetCode(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
	hashString := hex.EncodeToString(hash)
	cache.funcLock.Lock()
	defer cache.funcLock.Unlock()
	if _, exists := cache.functions[hashString]; !exists {
		//TODO: use this error! should be used!
		codeStored, _ := GetMethodCode(cache.api, hashString)
		cache.functions[hashString] = *codeStored
	}
	localCode := cache.functions[hashString]
	ops, blocks := parseBytes(localCode.CodeBytes)

	funcType := FunctionType{
		params:  localCode.CodeParams,
		results: localCode.CodeResults,
		string:  hex.EncodeToString(hash), //so you can lie better.
	}
	return funcType, ops, blocks
}
func (cache *DBCache) PreCacheCode(hash []byte) error {
	hashString := hex.EncodeToString(hash)
	cache.funcLock.Lock()
	defer cache.funcLock.Unlock()
	if _, exists := cache.functions[hashString]; exists {
		//we have it locally already!
		return nil
	}
	codeStored, err := GetMethodCode(cache.api, hashString)
	if err != nil {
		return err
	}
	cache.functions[hashString] = *codeStored
	return nil
}
func (cache *DBCache) GetContract(address string) (*Contract, error) {
	cache.conLock.Lock()
	defer cache.conLock.Unlock()
	contract, exists := cache.contracts[address]
	if exists {
		return contract, nil
	}
	//hopefully we just don't have it locally.
	contract, err := GetContractData(cache.api, address)
	if err != nil {
		return nil, err
	}
	cache.contracts[address] = contract
	return contract, nil
}
func (cache *DBCache) PreCacheContract(address string) error {
	cache.conLock.Lock()
	defer cache.conLock.Unlock()
	if _, exists := cache.contracts[address]; exists {
		return nil //we already have it locally
	}
	if contract, err := GetContractData(cache.api, address); err != nil {
		return err
	} else {
		cache.contracts[address] = contract
	}
	return nil
}
func (cache *DBCache) UploadChangedContracts() error {
	// UploadContract(cache.api)
	return nil
}

// API DB Spoofing
type DBSpoofer struct {
	storedFunctions map[string]CodeStored //hash=>functions
	storedContracts map[string]*Contract
}

func NewDBSpoofer() *DBSpoofer {
	return &DBSpoofer{
		storedFunctions: map[string]CodeStored{},
		storedContracts: map[string]*Contract{},
	}
}
func (spoof *DBSpoofer) GetContract(contractAddress string) (*Contract, error) {
	if contract, exists := spoof.storedContracts[contractAddress]; exists {
		return contract, nil
	} else {
		return nil, fmt.Errorf("contract not stored")
	}
}
func (spoof *DBSpoofer) AddContract(address string, contract *Contract) {
	spoof.storedContracts[address] = contract
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

func (spoof *DBSpoofer) AddSpoofedCode(hash string, funcCode CodeStored) {
	spoof.storedFunctions[hash] = funcCode
}
func (spoof *DBSpoofer) AddModuleToSpoofedCode(input interface{}) ([][]byte, error) {
	var mod Module
	switch v := input.(type) {
	case Module:
		mod = v
	case string:
		hexBinary, err := hex.DecodeString(v)
		if err != nil {
			return nil, err
		}
		mod = *decode(hexBinary)
	case []byte:
		mod = *decode(v)
	case []string:
		hashes := [][]byte{}
		for i := 0; i < len(v); i++ {
			foo, err := spoof.AddModuleToSpoofedCode(v[i])
			if err != nil {
				return nil, err
			}
			for i := 0; i < len(foo); i++ {
				hashes = append(hashes, foo[i])
			}
		}
		return hashes, nil
	case [][]byte:
		hashes := [][]byte{}
		for i := 0; i < len(v); i++ {
			foo, err := spoof.AddModuleToSpoofedCode(v[i])
			if err != nil {
				return nil, err
			}
			for i := 0; i < len(foo); i++ {
				hashes = append(hashes, foo[i])
			}
		}
		return hashes, nil
	}
	hashes := [][]byte{}
	for x := range mod.functionSection {
		code := CodeStored{
			CodeParams:  mod.typeSection[x].params,
			CodeResults: mod.typeSection[x].results,
			CodeBytes:   mod.codeSection[x].body,
		}
		localHash, err := code.Hash()

		if err != nil {
			return nil, err
		}
		spoof.AddSpoofedCode(hex.EncodeToString(localHash), code)
		hashes = append(hashes, localHash)
	}
	return hashes, nil
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

//GENERALLY USEFUL

func addressToInts(address interface{}) []uint64 {
	//if an empty address is returned, it should take the same amount of space as a full one.
	//converts and address to an array of uint64s, that way it can be pushed to the stack with more ease
	ans := []uint64{0, 0, 0, 0}
	var addressBytes []byte
	switch v := address.(type) {
	case common.Address:
		addressBytes = v.Bytes()
	case []byte:
		if len(v) != common.AddressLength { //224 bits
			return addressToInts(common.BytesToAddress(v))
		}
		addressBytes = v
	}
	for i := 0; i < 8-common.AddressLength%8; i++ {
		//we've made sure that the address is the correct address length.
		//Now, we need to make sure that it is divisible into uint64s.
		addressBytes = append(addressBytes, 0)
	}
	for i := 0; len(addressBytes) >= 8; i++ {
		ans[i] = LE.Uint64(addressBytes[:8])
		addressBytes = addressBytes[8:]
	}

	return ans
}
func uintsArrayToAddress(input []uint64) []byte {
	ans := []byte{}
	for i := 0; i < len(input); i++ {
		ans = LE.AppendUint64(ans, input[i])
	}
	return ans[:common.AddressLength]
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
		hashA, err := a.Code[i].Hash()
		hashB, errB := b.Code[i].Hash()
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

func (code CodeStored) Hash() ([]byte, error) {
	packedData, err := msgpack.Marshal(&code)
	if err != nil {
		return nil, err
	}
	hasher := md5.New()
	hasher.Write(packedData)
	return hasher.Sum(nil), nil
}
func contractDataToContract(cdata ContractData) *Contract {
	con := Contract{
		Address: common.HexToAddress(cdata.Address),
		Storage: cdata.Storage,
	}
	for _, code := range cdata.Methods {
		foo, _ := hex.DecodeString(code)
		con.CodeHashes = append(con.CodeHashes, hex.EncodeToString(foo))
	}
	return &con
}
func contractToContractData(con Contract) ContractData {
	cdata := ContractData{
		Address: con.Address.Hex(),
		Methods: []string{},
		Storage: con.Storage,
	}
	for _, code := range con.Code {
		foo, _ := code.Hash()
		cdata.Methods = append(cdata.Methods, hex.EncodeToString(foo))
	}
	return cdata
}

func (con Contract) Hash() common.Hash {
	contractBytes, err := ContractToMSGPackBytes(con)
	if err != nil {
		return common.BytesToHash([]byte{0})
	}
	h := sha256.New()
	h.Write(contractBytes)
	// fmt.Printf("%x", h.Sum(nil))
	return common.BytesToHash(h.Sum(nil))
}

func ContractToMSGPackBytes(con Contract) ([]byte, error) {
	cdata := contractToContractData(con)

	packedData, err := msgpack.Marshal(&cdata)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return packedData, nil
}

func DecodeModule(moduleBytes []byte) Module {
	l := *decode(moduleBytes)
	return l
}

func ModuleToCodeStored(m *Module) []CodeStored {
	cs := []CodeStored{}
	for i := 0; i < len(m.functionSection); i++ {
		cs = append(cs, CodeStored{
			CodeParams:  m.typeSection[i].params,
			CodeResults: m.typeSection[i].results,
			CodeBytes:   m.codeSection[i].body,
		})
	}

	return cs
}
