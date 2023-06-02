package cmd

import (
	"testing"

	"github.com/abiosoft/ishell/v2"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
)

func TestCaesarMessaging(t *testing.T) {
	nw := NewNetWorker()
	accountH := NewAccountHandler()
	seedNode := NewSeedHandler()
	seedShell := ishell.New()
	seedShell.AddCmd(seedNode.GetSeedCommands())
	seedShell.Process("seed")
	seedString := seedNode.hosting.GetConnectionString()

	a := NewCaesarHandler(accountH, nw)
	aUser, _ := accounts.GenerateAccount()
	a.accounts.AddAccountByAccount(*aUser)
	aShell := ishell.New()
	aShell.AddCmd(a.GetCommands())
	aShell.Process("caesar", "start", seedString)
	b := NewCaesarHandler(accountH, nw)
	bUser, _ := accounts.GenerateAccount()
	a.accounts.AddAccountByAccount(*bUser)
	bShell := ishell.New()
	bShell.AddCmd(b.GetCommands())
	bShell.Process("caesar", "start", seedString)

	aShell.Process("caesar", "talk", crypto.B58encode(bUser.PublicKey), "hello!")
	bShell.Process("caesar", "talk", crypto.B58encode(aUser.PublicKey), "how are you?")
	aLogs := a.chatLogs[bUser.Address]
	bLogs := b.chatLogs[aUser.Address]
	for i, aMsg := range aLogs {
		bMsg := bLogs[i]
		assert.Equal(t, aMsg.text, bMsg.text, "msgs did not have same text")
	}
}
