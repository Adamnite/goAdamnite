package main

import (
	"github.com/abiosoft/ishell/v2"
	"github.com/adamnite/go-adamnite/app/cmd"
)

func main() {

	shell := ishell.New()
	accountHandler := cmd.NewAccountHandler()
	shell.AddCmd(accountHandler.GetAccountCommands())

	msgHandling := cmd.NewCaesarHandler(accountHandler)
	shell.AddCmd(msgHandling.GetCaesarCommands())

	seedHandling := cmd.NewSeedHandler()
	shell.AddCmd(seedHandling.GetSeedCommands())

	shell.Interrupt(func(c *ishell.Context, count int, input string) {
		if count == 2 {
			msgHandling.Stop(c)
			seedHandling.Stop(c)
			shell.Close()
			return
		}
		if msgHandling.HoldingFocus {
			msgHandling.HoldingFocus = false
			c.Println("exiting chat")
			shell.SetPrompt(">adm>")
			shell.Stop()
			shell.Run()
		}

	})
	shell.SetPrompt(">adm>")
	shell.Println("Welcome to Adamnite!")

	// run shell
	shell.Run()
}
