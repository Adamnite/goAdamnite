package main

import (
	"github.com/adamnite/go-adamnite/internal/node"
	"github.com/adamnite/go-adamnite/internal/utils"
	"github.com/urfave/cli/v2"
)

type gniteConfig struct {
	Node node.Config
}

func defaultGniteNodeConfig() node.Config {
	cfg := node.DefaultConfig

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
		utils.Fatalf("Failed to create an Adamnite node: %v", err)
	}

	return gniteNode
}
