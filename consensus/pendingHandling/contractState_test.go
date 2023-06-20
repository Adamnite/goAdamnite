package pendingHandling

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

func TestContractStates(t *testing.T) {
	accountGoal := 5

	transactions := []*utils.Transaction{}
	db := VM.NewSpoofedDBCache(nil, nil)
	contract := VM.Contract{
		Address:    common.HexToAddress("0x123456"),
		CodeHashes: []string{},
	}
	methodsAsString := "" //TODO: get a module to use
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
	for i := 0; i < accountGoal; i++ {
		sender, _ := accounts.GenerateAccount()
		recipient, _ := accounts.GenerateAccount()

		tr, _ := utils.NewTransaction(sender, recipient.GetAddress(), big.NewInt(0), big.NewInt(1000))
		transaction, err := utils.NewVMTransaction(sender, tr, []byte{})
		if err != nil {
			t.Fatal(err)
		}
		transactions = append(transactions, transaction)
	}

}
