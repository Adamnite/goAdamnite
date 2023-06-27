package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/abiosoft/ishell/v2"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

type NetWorker struct {
	server         *networking.NetNode
	networkAccount *accounts.Account
	startedAt      time.Time
}

func NewNetWorker() *NetWorker {
	nw := NetWorker{}

	return &nw
}

func (nw *NetWorker) StartNetWorker() {
	if nw.server != nil {
		return
	}
	nw.startedAt = time.Now()
	if nw.networkAccount == nil {
		nw.networkAccount, _ = accounts.GenerateAccount()
	}
	nw.server = networking.NewNetNode(nw.networkAccount.Address)
}

func (nw *NetWorker) GetNetNode() *networking.NetNode {
	if nw.server == nil {
		nw.StartNetWorker()
	}
	return nw.server
}

func (nw *NetWorker) GetCommands() *ishell.Cmd {
	net := ishell.Cmd{
		Name: "net",
		Help: "networking is shared between all running node types, and is the interface of the P2P network",
	}
	net.AddCmd(&ishell.Cmd{
		Name:     "connect",
		Help:     "connect <to> : to should be the endpoint with a server running",
		LongHelp: "connect <to> : to should be the endpoint with a server running. If no <to> is passed, will use a prompt",
		Func:     nw.ConnectTo,
	})
	net.AddCmd(&ishell.Cmd{
		Name: "stats",
		Help: "get stats on the network functionality and change settings",
		Func: nw.GetStats,
	})

	return &net
}

func (nw *NetWorker) ConnectTo(c *ishell.Context) {
	if nw.server == nil {
		nw.StartNetWorker()
	}
	var conPoint string
	if len(c.Args) == 0 {
		c.Println("no connection string was passed, would you like to enter one manually? (leave blank to ignore)")
		c.Print("connect to: ")
		c.ShowPrompt(false)
		conPoint = c.ReadLine()
		if conPoint == "" || conPoint == "\n" {
			c.Println("canceling now")
			return
		}
		c.ShowPrompt(true)
	} else {
		conPoint = c.Args[0]
	}
	err := nw.server.ConnectToSeed(conPoint)
	if err != nil {
		c.Print("error connecting to that node, with error: ")
		c.Println(err)
	} else {
		c.Println("connected to point!")
	}
}

func (nw *NetWorker) GetStats(c *ishell.Context) {
	if nw.server == nil {
		nw.StartNetWorker()
	}
	//displaying only
	dif := time.Since(nw.startedAt).Round(time.Second / 100)
	var startTime string
	if dif.Hours() < 24 {
		startTime = nw.startedAt.Format("15:04:05") //TODO:go 1.20 has this as Time.TimeOnly, if we update, change these!
	} else {
		startTime = nw.startedAt.Format("2006-01-02 15:04:05") //go 1.20 time.DateTime
	}
	displaying := []string{
		"exit",
		fmt.Sprintf("Started at: %s running for: %s", startTime, dif.String()),
		fmt.Sprintf("Hosting from: %s", nw.server.GetOwnContact().ConnectionString),
	}
	//selectable values
	selectable := []string{
		fmt.Sprintf(
			"current connections: %v \t%%%3.3f of open connections are full\n\tmax connections: %v",
			nw.server.GetActiveConnectionsCount(),
			(float64(nw.server.GetActiveConnectionsCount()) / float64(nw.server.MaxOutboundConnections)),
			nw.server.MaxOutboundConnections,
		), //select to change max connection
	}
	if nw.server.GetGreylistSize() != 0 && nw.server.GetMaxGreylist() != 0 {
		selectable = append(selectable, fmt.Sprintf(
			"greylist has %v nodes. %%%3.3f of max(%v)",
			nw.server.GetGreylistSize(),
			float64(nw.server.GetGreylistSize())/float64(nw.server.GetMaxGreylist()),
			float64(nw.server.GetMaxGreylist()),
		))
	} else {
		selectable = append(selectable, fmt.Sprintf(
			"greylist has %v nodes. max(%v)",
			nw.server.GetGreylistSize(),
			float64(nw.server.GetMaxGreylist()),
		))
	}

	//methods (i think these might all get move to their own methods only)
	methods := []string{
		"get more connections", //sprawls out the connections
		"cutoff lowest percent",
	}

	//displaying
	options := append(displaying, selectable...)
	options = append(options, methods...)
	choice := c.MultiChoice(options, "Server Stats")
	c.Println(choice)
	if choice >= len(displaying) && choice-len(displaying) < len(selectable) {
		//selectable
		switch choice - len(displaying) {
		case 0:
			c.Print("Set new max node connections: ")
			c.ShowPrompt(false)
			newVal := c.ReadLine()
			c.ShowPrompt(true)

			if i, err := strconv.Atoi(newVal); err != nil {
				c.Println("sorry, i couldn't get that as a number")
			} else {
				c.Println("Setting new max connections to ", uint(i))
				nw.server.SetMaxConnections(uint(i))
			}
		case 1:
			c.Print("Set the new max greylist to(0 to not have a max): ")
			c.ShowPrompt(false)
			newVal := c.ReadLine()
			c.ShowPrompt(true)

			if i, err := strconv.Atoi(newVal); err != nil {
				c.Println("sorry, i couldn't get that as a number")
			} else {
				c.Println("Setting new max connections to ", uint(i))
				nw.server.SetMaxGreyList(uint(i))
			}
		default:
			c.Println("hi")

		}
	} else if choice >= len(selectable) {
		switch choice - len(selectable) - len(displaying) {
		case 0:
			nw.SprawlConnections(c)
		}
		//method
	} else {
		//displaying items
		return
	}
}
func (nw *NetWorker) SprawlConnections(c *ishell.Context) {
	if nw.server == nil {
		nw.StartNetWorker()
	}
	var layers string
	var cutoff string
	if len(c.Args) == 2 { //no gui
		layers = c.Args[0]
		cutoff = c.Args[1]
	} else {
		c.Println("Sprawling the network")
		c.ShowPrompt(false)
		c.Print("layers to call: ")
		layers = c.ReadLine()
		c.Print("cutoff percentage (0-1): ")
		cutoff = c.ReadLine()
		c.ShowPrompt(true)
	}
	layerVal, err := strconv.Atoi(layers)
	if err != nil {
		c.Println("error parsing layer value: ", err)
		return
	}
	cutoffVal, err := strconv.ParseFloat(cutoff, 32)
	if err != nil {
		c.Println("error parsing cutoff value: ", err)
		return
	}
	c.ProgressBar().Start()
	c.ProgressBar().Indeterminate(true)

	if err := nw.server.SprawlConnections(layerVal, float32(cutoffVal)); err != nil {
		c.Println("error sprawling: ", err)
		c.ProgressBar().Stop()
		return
	}
	c.ProgressBar().Stop()
	c.Println("network has been expanded!")
}
