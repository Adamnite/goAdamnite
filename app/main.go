package main

import (
	"github.com/abiosoft/ishell/v2"
	"github.com/adamnite/go-adamnite/app/cmd"
)

type handlers struct {
	accounts  *cmd.AccountHandler
	caesar    *cmd.CaesarHandler
	seeds     *cmd.SeedHandler
	consensus *cmd.ConsensusHandler
}

func getHandlers() *handlers {
	h := handlers{
		accounts: cmd.NewAccountHandler(),
	}
	h.caesar = cmd.NewCaesarHandler(h.accounts)
	h.consensus = cmd.NewConsensusHandler(*h.accounts)
	return &h
}
func (h *handlers) setHandlerCmds(shell *ishell.Shell) {
	shell.AddCmd(h.accounts.GetAccountCommands())
	shell.AddCmd(h.caesar.GetCaesarCommands())
	shell.AddCmd(h.consensus.GetConsensusCommands())
}

func main() {

	shell := ishell.New()
	handlers := getHandlers()
	handlers.setHandlerCmds(shell)

	seedHandling := cmd.NewSeedHandler()
	shell.AddCmd(seedHandling.GetSeedCommands())

	shell.Interrupt(func(c *ishell.Context, count int, input string) {
		if count == 2 {
			handlers.caesar.Stop(c)
			handlers.seeds.Stop(c)
			handlers.consensus.Stop(c)
			shell.Close()
			return
		}
		if handlers.caesar.HoldingFocus {
			handlers.caesar.HoldingFocus = false
			c.Println("exiting chat")
			shell.SetPrompt(">adm>")
			shell.Stop()
			shell.Run()
		}
	})
	shell.SetPrompt(">adm>")
	shell.Println("Welcome to Adamnite! You can try 'help' to see available commands")

	// run shell
	shell.Run()
}
