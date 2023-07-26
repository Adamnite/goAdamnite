package pendingHandling

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
)

func TestPendingTransactions(t *testing.T) {
	tq := NewQueue(false)
	sender, _ := accounts.GenerateAccount()
	recipient, _ := accounts.GenerateAccount()
	testTransaction, _ := utils.NewBaseTransaction(sender, recipient.Address, big.NewInt(1), big.NewInt(1))
	tq.AddToQueue(testTransaction)
	tq.Remove(testTransaction)
	ans := tq.Pop()
	assert.Nil(t, ans, "transaction is queued to be removed. Should return nil")
	testTransaction, _ = utils.NewBaseTransaction(sender, recipient.Address, big.NewInt(1), big.NewInt(1))
	tq.AddToQueue(testTransaction)
	ans = tq.Pop()

	assert.NotNil(t, ans, "nothing returned")
	assert.True(t, ans.Equal(testTransaction), "transaction not equal after being returned")
	if tq.Pop() != nil {
		fmt.Println("popped more transactions than it should have")
		t.Fail()
	}
}

func TestSorting(t *testing.T) {
	tq := NewQueue(false)
	sender, _ := accounts.GenerateAccount()
	recipient, _ := accounts.GenerateAccount()
	transactions := []*utils.BaseTransaction{}
	for i := 0; i < 50; i++ {
		testTransaction, _ := utils.NewBaseTransaction(sender, recipient.Address, big.NewInt(int64(i)), big.NewInt(int64(i)))
		testTransaction.Time = testTransaction.Time.Add(time.Duration(i) * time.Second * -1) //subtract the point I
		transactions = append(transactions, testTransaction)
		tq.AddToQueue(testTransaction)
	}
	tq.Remove(transactions[5])
	tq.SortQueue()
	if !tq.Pop().Equal(transactions[len(transactions)-1]) {
		fmt.Println("popping the wrong value after sorting")
		t.Fail()
	}
	RemoveAllFrom(transactions, tq)

	if tq.Pop() != nil {
		fmt.Println("popping found non existent transaction")
		t.Fail()
	}
	vagueTransactions := []utils.TransactionType{}
	for i := 0; i < 50; i++ {
		testTransaction, _ := utils.NewBaseTransaction(sender, recipient.Address, big.NewInt(int64(i)), big.NewInt(int64(i)))
		testTransaction.Time = testTransaction.Time.Add(time.Duration(i) * time.Second * -1) //subtract the point I
		vagueTransactions = append(vagueTransactions, testTransaction)
		tq.AddToQueue(testTransaction)
	}
	assert.Equal(
		t, len(vagueTransactions), tq.pendingQueue.Len(),
		"adding by interface did not load",
	)
	assert.Equal(
		t, vagueTransactions[0], tq.Pop(),
		"vague transactions not the same after being popped from queue",
	)
	tq.RemoveAll(vagueTransactions)
	if tq.Pop() != nil {
		fmt.Println("popping found non existent transaction")
		t.Fail()
	}
}
