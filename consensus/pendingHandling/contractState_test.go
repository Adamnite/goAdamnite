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
)

var (
	state   *statedb.StateDB
	senders []*accounts.Account
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
}
func TestContractStates(t *testing.T) {
	setup(1, big.NewInt(5))

	transactions := []*utils.VMCallTransaction{}
	db := VM.NewSpoofedDBCache(nil, nil)
	contract := VM.Contract{
		Address:    common.HexToAddress("0x123456"),
		CodeHashes: []string{},
	}
	methodsAsString := "0061736d01000000010a0260027f7f017f600000030302000105030100020614037f01419088040b7f004180080b7f004188080b073705066d656d6f727902000367657400001648656c6c6f576f726c645f44656661756c7443544f52000105626c6f636b0301036d736703020a0c020700200120006a0b02000b0b1301004180080b0c0000000000000000040400000045046e616d65011e020003676574011648656c6c6f576f726c645f44656661756c7443544f52071201000f5f5f737461636b5f706f696e746572090a0100072e726f64617461" //TODO: get a module to use
	err, hashes := db.DB.AddModuleToSpoofedCode(methodsAsString)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range hashes {
		contract.CodeHashes = append(contract.CodeHashes, hex.EncodeToString(x))
	}
	db.DB.AddContract(contract.Address.Hex(), &contract)

	csh, _ := NewContractStateHolder("ignoredAPIEndpoint")
	csh.dbCache = db
	for _, sender := range senders {

		tr, _ := utils.NewBaseTransaction(sender, contract.Address, big.NewInt(0), big.NewInt(1000))
		transaction, err := utils.NewVMTransactionFrom(sender, tr, append(hashes[0], 0x7f, 0x1, 0x7f, 0x2))
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

	if err := csh.RunAll(state); err != nil {
		t.Fatal(err)
	}

}
