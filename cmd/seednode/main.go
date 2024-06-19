// seednode runs a seednode for the Adamnite Gossip(P2P) Discovery Protocol

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/adamnite/go-adamnite/log"
	"github.com/libp2p/go-libp2p"
)

func main() {
	var (
		// listenAddr = flag.String("addr", ":39008", "listen address for Adamnite node")
		verbosity = flag.Int("verbosity", 3, "log verbosity (0-5)")
		vmodule   = flag.String("vmodule", "", "log verbosity pattern")
	)
	flag.Parse()

	glogger := log.NewGlogHandler(log.NewTerminalHandler(os.Stderr, false))
	slogVerbosity := log.FromLegacyLevel(*verbosity)
	glogger.Verbosity(slogVerbosity)
	glogger.Vmodule(*vmodule)
	log.SetDefault(log.NewLogger(glogger))

	// start a libp2p node with default settings
	node, err := libp2p.New()
	if err != nil {
		panic(err)
	}

	// print the node's listening addresses
	fmt.Println("Listen addresses:", node.Addrs())

	// shut the node down
	if err := node.Close(); err != nil {
		panic(err)
	}
}
