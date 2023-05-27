package main

import (
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/adamnite/go-adamnite/app/cmd"
)

func main() {

	shell := ishell.New()

	msgHandling := cmd.NewCaesarHandler()
	shell.AddCmd(msgHandling.GetCaesarCommands())
	shell.SetPrompt(">adm>")

	shell.Run()
}
