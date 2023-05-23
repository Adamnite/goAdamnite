package consensus

import (
	"encoding/hex"
	"log"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/common"
)

var (
	apiEndpoint         = "http://127.0.0.1:5000/"
	addTwoFunctionCode  = "0061736d0100000001170460027e7e017e60017e017e60017e017f60027e7e017f032120000000000000000000000000000000010101010101020303030303030303030307eb0120036164640000037375620001036d756c0002056469765f730003056469765f7500040572656d5f7300050572656d5f75000603616e640007026f72000803786f7200090373686c000a057368725f73000b057368725f75000c04726f746c000d04726f7472000e03636c7a000f0363747a001006706f70636e74001109657874656e64385f7300120a657874656e6431365f7300130a657874656e6433325f7300140365717a00150265710016026e650017046c745f730018046c745f750019046c655f73001a046c655f75001b0467745f73001c0467745f75001d0467655f73001e0467655f75001f0af301200700200020017c0b0700200020017d0b0700200020017e0b0700200020017f0b070020002001800b070020002001810b070020002001820b070020002001830b070020002001840b070020002001850b070020002001860b070020002001870b070020002001880b070020002001890b0700200020018a0b05002000790b050020007a0b050020007b0b05002000c20b05002000c30b05002000c40b05002000500b070020002001510b070020002001520b070020002001530b070020002001540b070020002001570b070020002001580b070020002001550b070020002001560b070020002001590b0700200020015a0b00f401046e616d6502ec012000020001780101790102000178010179020200017801017903020001780101790402000178010179050200017801017906020001780101790702000178010179080200017801017909020001780101790a020001780101790b020001780101790c020001780101790d020001780101790e020001780101790f0100017810010001781101000178120100017813010001781401000178150100017816020001780101791702000178010179180200017801017919020001780101791a020001780101791b020001780101791c020001780101791d020001780101791e020001780101791f02000178010179"
	addTwoFunctionBytes = []byte{}
	addTwoCodeStored    VM.CodeStored
	addTwoFunctionHash  []byte
	testContract        VM.Contract
	testAccount         = common.Address{0, 1, 2}
)

func setup() {
	addTwoFunctionBytes, _ = hex.DecodeString(addTwoFunctionCode)
	mod := VM.DecodeModule(addTwoFunctionBytes)
	stored, _, err := VM.UploadModuleFunctions(apiEndpoint, mod)
	if err != nil {
		panic(err)
	}

	addTwoCodeStored = stored[0]
	addTwoFunctionHash, _ = addTwoCodeStored.Hash()
	testContract = VM.Contract{CallerAddress: common.Address{1}, Value: big.NewInt(0), Input: nil, Gas: 30000, CodeHashes: []string{hex.EncodeToString(addTwoFunctionHash)}}

}
func TestProcessingRun(t *testing.T) {
	setup()
	bNode, err := NewBConsensus(apiEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VM.UploadMethod(apiEndpoint, addTwoCodeStored); err != nil {
		if err == VM.ErrConnectionRefused {
			t.Log("server is not running. Try again with the Offchain DB running")
			t.Skip(err)
		}
		t.Fatal(err)
	}
	if err := VM.UploadContract(apiEndpoint, testContract); err != nil {
		t.Fatal(err)
	}
	claim := VM.RuntimeChanges{
		Caller:            testAccount,
		CallTime:          time.Now().UTC(),
		ContractCalled:    testContract.Address,
		ParametersPassed:  append(addTwoFunctionHash, byte(VM.Op_i64)),
		GasLimit:          10000,
		ChangeStartPoints: []uint64{0},
		Changed:           [][]byte{{0, 1, 2, 3, 4, 5, 6, 7}},
	}
	//set the parameters to the hash, types, and values
	claim.ParametersPassed = append(claim.ParametersPassed, VM.EncodeUint64(1)...)
	claim.ParametersPassed = append(claim.ParametersPassed, VM.Op_i64)
	claim.ParametersPassed = append(claim.ParametersPassed, VM.EncodeUint64(2)...)
	didPass, _, err := bNode.VerifyRun(claim)
	if didPass {
		log.Println("should not pass, used incomplete claim")
		t.Fatal(err)
	}
	err = bNode.ProcessRun(&claim)
	if err != nil {
		t.Fatal(err)
	}
	didPass, _, err = bNode.VerifyRun(claim)
	if !didPass {
		log.Println("failed to get same results running same test twice")
		t.Fatal(err)
	}

}
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
