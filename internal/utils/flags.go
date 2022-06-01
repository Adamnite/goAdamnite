package utils

import (
	"io/ioutil"
	"strings"

	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/node"
	"github.com/adamnite/go-adamnite/p2p"
	"github.com/adamnite/go-adamnite/p2p/enode"
	"github.com/adamnite/go-adamnite/params"
	"github.com/urfave/cli/v2"
)

var (
	PasswordFileFlag = cli.StringFlag{
		Name:  "password",
		Usage: "Password file to use for non-interactive password input",
		Value: "",
	}
)

func SetNodeConfig(ctx *cli.Context, cfg *node.Config) {
	SetP2PConfig(ctx, &cfg.P2P)
}

func SetP2PConfig(ctx *cli.Context, cfg *p2p.Config) {
	setBootstrapNode(ctx, cfg)
}

func setBootstrapNode(ctx *cli.Context, cfg *p2p.Config) {
	urls := params.MainnetBootnodes

	cfg.BootstrapNodes = make([]*enode.Node, 0, len(urls))
	for _, url := range urls {
		if url != "" {
			node, err := enode.Parse(enode.ValidSchemes, url)
			if err != nil {
				log15.Crit("Bootstrap URL invalid", "enode", url, "err", err)
				continue
			}
			cfg.BootstrapNodes = append(cfg.BootstrapNodes, node)
		}
	}
}

func MigrateFlags(action func(ctx *cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				ctx.Set(name, ctx.String(name))
			}
		}
		return action(ctx)
	}
}

// MakePasswordList reads password lines from the file specified by the global --password flag.
func MakePasswordList(ctx *cli.Context) []string {
	path := ctx.String(PasswordFileFlag.Name)
	if path == "" {
		return nil
	}
	text, err := ioutil.ReadFile(path)
	if err != nil {
		Fatalf("Failed to read password file: %v", err)
	}
	lines := strings.Split(string(text), "\n")
	// Sanitise DOS line endings.
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines
}
