package cmd

import (
	"strconv"

	"github.com/abiosoft/ishell/v2"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/networking"
)

// this is for handling any seed types to be started through the CLI
type SeedHandler struct {
	hosting *networking.NetNode
}

func NewSeedHandler() *SeedHandler {
	return &SeedHandler{}
}
func (sh *SeedHandler) GetSeedCommands() *ishell.Cmd {
	seedFuncs := ishell.Cmd{
		Name: "seed",
		Help: "seed node. Basically a blank server for connecting nodes together. Acts as the start point for others connections (this does the same as seed start)",
		Func: sh.Start,
	}
	seedFuncs.AddCmd(&ishell.Cmd{
		Name: "start",
		Help: "start <optionalBouncerPort> seed node server",
		Func: sh.Start,
	})
	seedFuncs.AddCmd(&ishell.Cmd{
		Name: "stop",
		Help: "stop shuts down the server",
		Func: sh.Stop,
	})
	seedFuncs.AddCmd(&ishell.Cmd{
		Name: "connectTo",
		Help: "connectTo <connectionString>, attempts to connect to the target point",
		Func: sh.ConnectTo,
	})

	return &seedFuncs
}
func (sh *SeedHandler) Start(c *ishell.Context) {
	if sh.hosting != nil {
		c.Println("server already started")
		return
	}
	sh.hosting = networking.NewNetNode(utils.Address{0})
	if err := sh.hosting.AddServer(); err != nil {
		c.Println(err)
		return
	}
	if len(c.Args) >= 1 {
		if i, err := strconv.Atoi(c.Args[0]); err != nil {
			c.Println("error parsing bouncer hosting port")
			return
		} else {
			sh.hosting.AddBouncerServer(nil, nil, uint32(i))
		}
	} else {
		sh.hosting.AddBouncerServer(nil, nil, 0)
	}

	c.Printf("Seed server has started up at %v\n", sh.hosting.GetConnectionString())
	c.Printf("Seed Bouncer Server available at %v\n", sh.hosting.GetBouncerString())
}
func (sh *SeedHandler) Stop(c *ishell.Context) {
	if sh.hosting == nil {
		c.Println("server already shut down")
		return
	}
	sh.hosting.Close()
	sh.hosting = nil
	c.Println("server has been stopped")
}
func (sh *SeedHandler) ConnectTo(c *ishell.Context) {
	if sh.hosting == nil {
		c.Println("server not running, starting it now")
		sh.Start(c)
	}
	if len(c.Args) == 0 {
		c.Println("no connection argument passed")
		return
	}
	c.Println("connecting to node point")
	if err := sh.hosting.ConnectToSeed(c.Args[0]); err != nil {
		c.Println("error connecting")
		c.Println(err)
		return
	}
	c.Println("connected successfully")
}
