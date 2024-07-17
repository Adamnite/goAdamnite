package main

import (
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

	gniteNode, err := node.New(&cfg.Node)
	if err != nil {
		log.Error("Failed to create an Adamnite node", "err", err)
	}

	return gniteNode
}
