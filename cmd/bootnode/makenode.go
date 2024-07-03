package main

import (
	"github.com/adamnite/go-adamnite/internal/node"
	"github.com/adamnite/go-adamnite/log"
	"github.com/urfave/cli/v2"
)

type gniteConfig struct {
	Node node.Config
}

func defaultBootNodeConfig() node.Config {
	cfg := node.DefaultBootNodeConfig

	return cfg
}

func loadBootNodeConfig(ctx *cli.Context) gniteConfig {
	cfg := gniteConfig{
		Node: defaultBootNodeConfig(),
	}

	return cfg
}

func makeBootNode(ctx *cli.Context) *node.Node {
	cfg := loadBootNodeConfig(ctx)

	bootNode, err := node.New(&cfg.Node)
	if err != nil {
		log.Error("Failed to create an Adamnite boot node", "err", err)
	}

	return bootNode
}
