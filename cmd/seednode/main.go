package main

import (
	"flag"
	"os"

	"github.com/adamnite/go-adamnite/log15"
)

func main() {
	var (
		// listenAddress = flag.String("addr", ":3880", "listen address")
		verbosity = flag.Int("verbosity", int(log15.LvlInfo), "log verbosity (0-5)")
		vmodule   = flag.String("vmodule", "", "log verbosity pattern")

		// nodeKey *ecdsa.PrivateKey
		// err     error
	)

	flag.Parse()

	glogger := log15.NewGlogHandler(log15.StreamHandler(os.Stderr, log15.TerminalFormat()))
	glogger.Verbosity(log15.Lvl(*verbosity))
	glogger.Vmodule(*vmodule)
	log15.Root().SetHandler(glogger)

	// db, err := enode.OpenDB("")
	// localNode := enode.NewLocalNode(db, nodeKey)
}
