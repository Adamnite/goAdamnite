package utils

import (
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/node"
	"github.com/adamnite/go-adamnite/p2p"
	"github.com/adamnite/go-adamnite/p2p/enode"
	"github.com/adamnite/go-adamnite/params"
	"github.com/urfave/cli/v2"
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
