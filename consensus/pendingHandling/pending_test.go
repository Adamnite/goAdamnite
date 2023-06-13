package pendingHandling

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

func TestPendingTransactions(t *testing.T) {
	tq := NewQueue()
	sender, _ := accounts.GenerateAccount()
	recipient, _ := accounts.GenerateAccount()
	testTransaction, _ := utils.NewTransaction(sender, recipient.Address, big.NewInt(1), big.NewInt(1))
	tq.AddToQueue(testTransaction)
	tq.Remove(testTransaction)
	testTransaction, _ = utils.NewTransaction(sender, recipient.Address, big.NewInt(1), big.NewInt(1))
	//we need to remake the testTransaction so it will have a different hash. This is done to prevent adding the same transaction to Queue twice
	tq.AddToQueue(testTransaction)
	if !tq.Pop().Equal(*testTransaction) {
		fmt.Println("the popped transaction isn't correct")
		t.Fail()
	}
	if tq.Pop() != nil {
		fmt.Println("popped more transactions than it should have")
		t.Fail()
	}
}

func TestSorting(t *testing.T) {
	tq := NewQueue()
	sender, _ := accounts.GenerateAccount()
	recipient, _ := accounts.GenerateAccount()
	transactions := []*utils.Transaction{}
	for i := 0; i < 50; i++ {
		testTransaction, _ := utils.NewTransaction(sender, recipient.Address, big.NewInt(int64(i)), big.NewInt(int64(i)))
		testTransaction.Time = testTransaction.Time.Add(time.Duration(i) * time.Second * -1) //subtract the point I
		transactions = append(transactions, testTransaction)
		tq.AddToQueue(testTransaction)
	}
	tq.Remove(transactions[5])
	tq.SortQueue()
	if !tq.Pop().Equal(*transactions[len(transactions)-1]) {
		fmt.Println("popping the wrong value after sorting")
		t.Fail()
	}
}
