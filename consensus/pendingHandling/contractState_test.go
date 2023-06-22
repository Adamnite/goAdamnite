package pendingHandling

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
)

var (
	state          *statedb.StateDB
	senders        []*accounts.Account
	db             *VM.SpoofedDBCache
	addTwoContract VM.Contract
	addTwoFunction []byte //the hash/name for the addTwoFunction
)

func setup(accountGoal int, startBalance *big.Int) {
	state, _ = statedb.New(common.Hash{}, statedb.NewDatabase(rawdb.NewMemoryDB()))
	for i := 0; i < accountGoal; i++ {
		ac, _ := accounts.GenerateAccount()
		senders = append(senders, ac)
		state.AddBalance(ac.GetAddress(), startBalance)
	}
	rootHash := state.IntermediateRoot(false)
	state.Database().TrieDB().Commit(rootHash, false, nil)

	db = VM.NewSpoofedDBCache(nil, nil)
	addTwoContract = VM.Contract{
		Address:    common.HexToAddress("0x123456"),
		CodeHashes: []string{},
	}
	methodsAsString := "0061736d01000000010a0260027f7f017f600000030302000105030100020614037f01419088040b7f004180080b7f004188080b073705066d656d6f727902000367657400001648656c6c6f576f726c645f44656661756c7443544f52000105626c6f636b0301036d736703020a0c020700200120006a0b02000b0b1301004180080b0c0000000000000000040400000045046e616d65011e020003676574011648656c6c6f576f726c645f44656661756c7443544f52071201000f5f5f737461636b5f706f696e746572090a0100072e726f64617461"
	//just has the add function.
	err, hashes := db.DB.AddModuleToSpoofedCode(methodsAsString)
	if err != nil {
		panic(err)
	}
	addTwoFunction = hashes[0]
	addTwoContract.CodeHashes = []string{hex.EncodeToString(addTwoFunction)}

	db.DB.AddContract(addTwoContract.Address.Hex(), &addTwoContract)

}
func verifySingleContractCallPerAccount(t *testing.T, numberOfAccounts int) {
	setup(numberOfAccounts, big.NewInt(5))

	transactions := []*utils.VMCallTransaction{}

	csh, _ := NewContractStateHolder("ignoredAPIEndpoint")
	csh.dbCache = db
	for _, sender := range senders {

		tr, _ := utils.NewBaseTransaction(sender, addTwoContract.Address, big.NewInt(0), big.NewInt(1000))
		transaction, err := utils.NewVMTransactionFrom(sender, tr, append(addTwoFunction, 0x7f, 0x1, 0x7f, 0x2))
		if err != nil {
			t.Fatal(err)
		}
		transactions = append(transactions, transaction)
	}

	for _, tr := range transactions {
		if err := csh.QueueTransaction(tr); err != nil {
			t.Fatal(err)
		}
	}
	assert.Equal(
		t, len(transactions), len(csh.contractsHeld[addTwoContract.Address.Hex()].transactions), "all transactions should be queued under same contract")

	if err := csh.RunAll(state); err != nil {
		t.Fatal(err)
	}

}

func TestOneAccountOneTransaction(t *testing.T) {
	verifySingleContractCallPerAccount(t, 1)
}
func TestMoreAccountsAndTransactions(t *testing.T) {
	verifySingleContractCallPerAccount(t, 100)
}
