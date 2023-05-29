package cmd

import (
	"testing"

	"github.com/abiosoft/ishell/v2"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/stretchr/testify/assert"
)

func TestCaesarMessaging(t *testing.T) {
	seedNode := NewSeedHandler()
	seedShell := ishell.New()
	seedShell.AddCmd(seedNode.GetSeedCommands())
	seedShell.Process("seed")
	seedString := seedNode.hosting.GetOwnContact().ConnectionString

	a := NewCaesarHandler()
	aShell := ishell.New()
	aShell.AddCmd(a.GetCaesarCommands())
	aShell.Process("caesar", "start", seedString)
	b := NewCaesarHandler()
	bShell := ishell.New()
	bShell.AddCmd(b.GetCaesarCommands())
	bShell.Process("caesar", "start", seedString)

	aShell.Process("caesar", "talk", crypto.B58encode(b.thisUser.PublicKey), "hello!")
	bShell.Process("caesar", "talk", crypto.B58encode(a.thisUser.PublicKey), "how are you?")
	aLogs := a.chatLogs[b.thisUser.Address]
	bLogs := b.chatLogs[a.thisUser.Address]
	for i, aMsg := range aLogs {
		bMsg := bLogs[i]
		assert.Equal(t, aMsg.text, bMsg.text, "msgs did not have same text")
	}
}
