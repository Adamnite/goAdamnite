package cmd

import (
	"github.com/abiosoft/ishell/v2"
	"github.com/adamnite/go-adamnite/consensus"
)

type ConsensusHandler struct {
	server   *consensus.ConsensusNode
	accounts AccountHandler
}

func NewConsensusHandler(ac AccountHandler) *ConsensusHandler {
	ch := ConsensusHandler{accounts: ac}
	return &ch
}
func (ch ConsensusHandler) isServerLive() bool {
	return ch.server != nil
}

func (ch *ConsensusHandler) GetConsensusCommands() *ishell.Cmd {
	conFuncs := ishell.Cmd{
		Name: "consensus",
		Help: "consensus handles communication of the chain, as well as storing data",
	}
	conFuncs.AddCmd(&ishell.Cmd{
		Name: "start",
		Help: "start starts the consensus handler",
		Func: ch.StartSelectable,
	})
	conFuncs.AddCmd(&ishell.Cmd{
		Name: "stop",
		Help: "stop shuts down the consensus handler server",
		Func: ch.Stop,
	})
	conFuncs.AddCmd(&ishell.Cmd{
		Name: "connect",
		Help: "connect <endpoint(or to interact, ignore)> attempts to connect to a node at that endpoint and get its connections",
		Func: ch.ConnectTo,
	})

	return &conFuncs
}

func (ch *ConsensusHandler) StartSelectable(c *ishell.Context) {
	if ch.isServerLive() {
		c.Println("server already running.")
	}
	severTypes := []string{
		"Chamber A\t|\tHandles direct transactions and user balances",
		"Chamber B\t|\tHandles VM interactions (as well as standard transactions) and needs its own OffChain DB to access smart contract states",
	}
	chamberSelection := c.MultiChoice(severTypes, "Select Consensus server type to run")
	c.Println("please select an account to sign/elect with")
	spenderAccount := ch.accounts.SelectAccount(c)
	if spenderAccount == nil {
		return
	}
	switch chamberSelection {
	case 0: //chamber A
		c.Println("Starting Chamber A server now!")
		server, err := consensus.NewAConsensus(*spenderAccount.account)
		if err != nil {
			c.Print("error starting: ")
			c.Println(err)
		}
		ch.server = server
	case 1: //chamber B
		c.Println("Starting Chamber B server now!")
		c.Println("sorry, not setup yet")
		return
		// server, err := consensus.NewBConsensus()
	}
	ch.ConnectTo(c)
	c.Println("server is online!")
}
func (ch *ConsensusHandler) Stop(c *ishell.Context) {
	if ch.server != nil {
		//TODO: properly close the server
		ch.server = nil
	}
	c.Println("consensus server closed")
}

func (ch *ConsensusHandler) ConnectTo(c *ishell.Context) {
	var seed string
	if len(c.Args) > 0 && c.Cmd.Name == "connect" {
		//we're being called directly with a parameter
		seed = c.Args[0]
	} else {
		c.Println("do you have a known seed node port to connect to?")
		c.Print("node@: ")
		c.ShowPrompt(false)
		seed = c.ReadLine()
		c.ShowPrompt(true)
	}
	if seed == "" {
		return
	}
	if len(seed) < 8 { //needs at least 4 decimal places
		c.Println("too short to be a connection string(should look like 8.8.8.8:1234)")
		return
	}
	if err := ch.server.ConnectTo(seed); err != nil {
		c.Print("err connecting: ")
		c.Println(err)
		return
	}
	c.Println("successfully connected!")
}
