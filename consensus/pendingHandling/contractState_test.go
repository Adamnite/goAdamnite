package pendingHandling

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
)

var (
	state             *statedb.StateDB
	senders           []*accounts.Account
	db                *VM.SpoofedDBCache
	addTwoContract    VM.Contract
	addTwoFunction    []byte //the hash/name for the addTwoFunction
	saveLocalFunction []byte
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
	hashes, err := db.DB.AddModuleToSpoofedCode(methodsAsString)
	if err != nil {
		panic(err)
	}
	addTwoFunction = hashes[0]
	// methodsAsString = "0061736d0100000001040160000003030200000405017001010105030100020608017f01418088040b070a01066d656d6f727902000a130202000b0e001080808080001080808080000b0036046e616d65011b02000564756d6d7901115f5f7761736d5f63616c6c5f64746f7273071201000f5f5f737461636b5f706f696e74657200760970726f647563657273010c70726f6365737365642d62790105636c616e675631342e302e33202868747470733a2f2f6769746875622e636f6d2f6c6c766d2f6c6c766d2d70726f6a656374203166393134303036346466626662306262646138653531333036656135313038306232663761616329"
	// err, hashes = db.DB.AddModuleToSpoofedCode(methodsAsString)
	// if err != nil {
	// 	panic(err)
	// }
	// saveLocalFunction = hashes[0]
	// addTwoContract.CodeHashes = []string{hex.EncodeToString(addTwoFunction), hex.EncodeToString(saveLocalFunction)}
	addTwoContract.CodeHashes = []string{hex.EncodeToString(addTwoFunction)}

	db.DB.AddContract(addTwoContract.Address.Hex(), &addTwoContract)

}
func verifySingleContractCallPerAccount(t *testing.T, numberOfAccounts int) {
	setup(numberOfAccounts, big.NewInt(5))

	transactions := []*utils.VMCallTransaction{}

	csh, _ := NewContractStateHolder(db)
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
func verifySingleContractBlockCallPerAccount(t *testing.T, numberOfAccounts int, blockSize int) {
	setup(numberOfAccounts, big.NewInt(500))

	transactions := []*utils.VMCallTransaction{}

	csh, _ := NewContractStateHolder(db)
	for _, sender := range senders {

		tr, _ := utils.NewBaseTransaction(sender, addTwoContract.Address, big.NewInt(0), big.NewInt(1000))
		transaction, err := utils.NewVMTransactionFrom(sender, tr, append(addTwoFunction, 0x7f, 0x1, 0x7f, 0x2))
		if err != nil {
			t.Fatal(err)
		}
		transactions = append(transactions, transaction)
	}
	timeBasedQuitter := make(chan (any))
	go func() {
		<-time.After(time.Second * 5)
		timeBasedQuitter <- struct{}{}
	}()

	for _, tr := range transactions {
		if err := csh.QueueTransaction(tr); err != nil {
			t.Fatal(err)
		}
	}

	finalTs, err := csh.RunOnUntil(state, blockSize, timeBasedQuitter, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t, len(transactions), len(finalTs), "all transactions should be queued under same contract")

	// if err := csh.RunAll(state); err != nil {
	// 	t.Fatal(err)
	// }
}
func TestOneContractCallBlock(t *testing.T) {
	verifySingleContractBlockCallPerAccount(t, 5, 5)
}

func TestOneAccountOneTransaction(t *testing.T) {
	verifySingleContractCallPerAccount(t, 1)
}
func TestMoreAccountsAndTransactions(t *testing.T) {
	verifySingleContractCallPerAccount(t, 100)
}
