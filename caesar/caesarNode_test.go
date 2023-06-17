package caesar

import (
	"math"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
)

func setupTestCaesarNode(autoConnectSeed *networking.Contact) (*accounts.Account, *CaesarNode) {
	account, _ := accounts.GenerateAccount()
	node := NewCaesarNode(account)
	node.Startup()
	if autoConnectSeed != nil {
		node.netHandler.ConnectToContact(autoConnectSeed)
	}
	return account, node
}

func TestTwoServersMessagingOpenly(t *testing.T) {
	seedNode := networking.NewNetNode(common.Address{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	seedNode.AddServer()
	seedContact := seedNode.GetOwnContact()

	aAccount, aNode := setupTestCaesarNode(&seedContact)

	bAccount, bNode := setupTestCaesarNode(&seedContact)
	seedNode.FillOpenConnections()
	testMessage, err := utils.NewCaesarMessage(bAccount, aAccount, "Hello World!")
	if err != nil {
		t.Fatal(err)
	}

	if err := aNode.SendMessage(testMessage); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1,
		len(bNode.msgByHash),
		"b node appears to have not received the message",
	)
	assert.Equal(t, 1,
		len(bNode.msgByRecipient),
		"b node appears to have not received the message",
	)
	assert.Equal(t, 1,
		len(bNode.msgBySender),
		"b node appears to have not received the message",
	)
}

func TestManyOpenMessages(t *testing.T) {
	seedNode := networking.NewNetNode(common.Address{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	seedNode.AddServer()
	seedContact := seedNode.GetOwnContact()

	testingAccounts := []*accounts.Account{}
	testingNodes := []*CaesarNode{}
	for i := 0; i < 5; i++ {
		a, b := setupTestCaesarNode(&seedContact)
		testingAccounts = append(testingAccounts, a)
		testingNodes = append(testingNodes, b)
	}
	for _, node := range testingNodes {
		node.netHandler.SprawlConnections(3, 0)
		node.netHandler.FillOpenConnections()
	}
	seedNode.FillOpenConnections()
	for _, node := range testingNodes {
		for _, target := range testingAccounts {
			if err := node.Send(target, "hi"); err != nil {
				t.Fatal("error from node:%w to: %w with err:%w", node, target, err)
			}
		}
	}

	for _, node := range testingNodes {
		//check everyone has the same messages, and all of them
		assert.Equal(t, math.Pow(float64(len(testingNodes)), 2), float64(len(node.msgByHash)), "not all messages seen")
	}
}
