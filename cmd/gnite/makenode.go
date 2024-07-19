package main

import (
	"github.com/adamnite/go-adamnite/internal/flags"
	"github.com/adamnite/go-adamnite/internal/node"
	"github.com/adamnite/go-adamnite/log"
	"github.com/urfave/cli/v2"
)

type gniteConfig struct {
	Node node.Config
}

func defaultGniteNodeConfig() node.Config {
	cfg := node.DefaultConfig

	cfg.ProtocolID = "/gnite/1.0.0"

	return cfg
}

func loadGniteConfig(ctx *cli.Context) gniteConfig {
	cfg := gniteConfig{
		Node: defaultGniteNodeConfig(),
	}

	return cfg
}

func makeGniteNode(ctx *cli.Context) *node.Node {
	cfg := loadGniteConfig(ctx)
	err := setParams(ctx, &cfg)
	if err != nil {
		log.Error("Failed to set params", "err", err)
	}

	gniteNode, err := node.New(&cfg.Node)
	if err != nil {
		log.Error("Failed to create an Adamnite node", "err", err)
	}

	return gniteNode
}

func setParams(ctx *cli.Context, cfg *gniteConfig) error {
	cfg.Node.P2PPort = uint32(ctx.Int(flags.NetworkPort.Name))

	if ctx.IsSet(flags.DataDir.Name) {
		cfg.Node.DataDir = ctx.String(flags.DataDir.Name)
	}
	return nil
}
