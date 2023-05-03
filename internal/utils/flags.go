package utils

import (
	"io/ioutil"
	"strings"

	"github.com/adamnite/go-adamnite/adm"
	"github.com/adamnite/go-adamnite/adm/adamconfig"
	"github.com/adamnite/go-adamnite/bargossip"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/fdutils"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/node"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/bargossip/nat"
	"github.com/adamnite/go-adamnite/params"
	"github.com/urfave/cli/v2"
)

var (
	PasswordFileFlag = cli.StringFlag{
		Name:  "password",
		Usage: "Password file to use for non-interactive password input",
		Value: "",
	}
	MainnetFlag = cli.BoolFlag{
		Name:  "mainnet",
		Usage: "Adamnite mainnet",
	}
	TestnetFlag = cli.BoolFlag{
		Name:  "testnet",
		Usage: "Adamnite testnet",
	}
	DemoFlag = cli.BoolFlag{
		Name:  "demo",
		Usage: "Adamnite POC demo version",
	}
	WitnessFalg = cli.BoolFlag{
		Name:  "witness",
		Usage: "Adamnite witness",
	}
	WitnessAddressFlag = cli.StringFlag{
		Name:  "witness.addr",
		Usage: "The address of witness",
	}
	NetworkIP = cli.StringFlag{
		Name:  "network.ip",
		Usage: "The external IP address of the node",
	}
	NATFlag = cli.StringFlag{
		Name:  "nat",
		Usage: "NAT port mapping mechanism (any|none|upnp|pmp|extip:<IP>)",
		Value: "any",
	}
)

func SetNodeConfig(ctx *cli.Context, cfg *node.Config) {
	SetP2PConfig(ctx, &cfg.P2P)
}

func SetAdamniteConfig(ctx *cli.Context, node *node.Node, cfg *adamconfig.Config) {
	switch {
	case ctx.Bool(MainnetFlag.Name):
		cfg.Genesis = core.DefaultGenesisBlock()
	case ctx.Bool(TestnetFlag.Name):
		cfg.Genesis = core.DefaultTestnetGenesisBlock()
	default:
		cfg.Genesis = core.DefaultTestnetGenesisBlock()
	}

	cfg.AdamniteDbHandles = MakeAdamniteDatabaseHandles()

	setWitnessAddress(ctx, cfg)
}

func setWitnessAddress(ctx *cli.Context, cfg *adamconfig.Config) {
	var witnessAddr string
	if ctx.IsSet(WitnessFalg.Name) && ctx.IsSet(WitnessAddressFlag.Name) {
		witnessAddr = ctx.String(WitnessAddressFlag.Name)
	}

	cfg.Validator.WitnessAddress = common.StringToAddress(witnessAddr)
}

func SetP2PConfig(ctx *cli.Context, cfg *bargossip.Config) {
	setBootstrapNode(ctx, cfg)
	setNAT(ctx, cfg)

	if ctx.String(NetworkIP.Name) != "" {
		cfg.ListenAddr = ctx.String(NetworkIP.Name)
	}
}

func setBootstrapNode(ctx *cli.Context, cfg *bargossip.Config) {
	urls := params.MainnetBootnodes

	cfg.BootstrapNodes = make([]*admnode.GossipNode, 0, len(urls))
	for _, url := range urls {
		if url != "" {
			nodeInfo, err := admnode.ParseNodeURL(url)
			if err != nil {
				log15.Crit("Bootstrap URL invalid", "admnode", url, "err", err)
				continue
			}

			node, err := admnode.New(nodeInfo)
			if err != nil {
				log15.Crit(err.Error())
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

func MakeAdamniteDatabaseHandles() int {
	hLimit, err := fdutils.GetMaxHandles()
	if err != nil {
		Fatalf("Failed to retrieve file descripto allowance: %v", err)
	}

	hRaised, err := fdutils.GetRaise(uint64(hLimit))
	if err != nil {
		Fatalf("Failed to raise file descriptor allowance: %v", err)
	}
	return int(hRaised / 2)
}

func RegisterAdamniteSerivce(node *node.Node, cfg *adamconfig.Config) *adm.AdamniteImpl {
	backend, err := adm.New(node, cfg)
	if err != nil {
		Fatalf("Failed to register the Adamnite service: %v", err)
	}

	return backend
}

// setNAT creates a port mapper from command line flags.
func setNAT(ctx *cli.Context, cfg *bargossip.Config) {
	if ctx.IsSet(NATFlag.Name) {
		natif, err := nat.Parse(ctx.String(NATFlag.Name))
		if err != nil {
			Fatalf("Option %s: %v", NATFlag.Name, err)
		}
		cfg.NAT = natif
	}
}
