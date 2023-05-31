package cmd

import (
	"github.com/abiosoft/ishell/v2"
	"github.com/adamnite/go-adamnite/consensus"
)

type ConsensusHandler struct {
	server *consensus.ConsensusNode
}

func NewConsensusHandler() *ConsensusHandler {
	ch := ConsensusHandler{}
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

	switch chamberSelection {
	case 0: //chamber A
		c.Println("Starting Chamber A server now!")
		// ch.server = consensus.
	}
}
