package cmd

import (
	"fmt"
	"testing"

	"github.com/abiosoft/ishell"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
)

func TestCaesarMessaging(t *testing.T) {
	seedNode := networking.NewNetNode(common.Address{0, 0, 0, 0})
	seedNode.AddServer()

	a := NewCaesarHandler()
	aShell := ishell.New()
	aShell.AddCmd(a.GetCaesarCommands())
	aShell.Process("caesar", "start", seedNode.GetOwnContact().ConnectionString)

	b := NewCaesarHandler()
	bShell := ishell.New()
	bShell.AddCmd(b.GetCaesarCommands())
	bShell.Process("caesar", "start", seedNode.GetOwnContact().ConnectionString)

	aShell.Process("caesar", "talk", crypto.B58encode(b.thisUser.PublicKey), "hello!")
	fmt.Println("\n\n\n ")
	bShell.Process("caesar", "talk", crypto.B58encode(a.thisUser.PublicKey), "how are you?")


}
