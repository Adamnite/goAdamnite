package caesar

import (
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
	// bAccount, _ := setupTestCaesarNode(&seedContact)
	seedNode.FillOpenConnections()
	testMessage, err := utils.NewCaesarMessage(*bAccount, *aAccount, "Hello World!")
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

}
