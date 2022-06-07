package adamconfig

import (
	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/node"
	"github.com/adamnite/go-adamnite/params"
)

type Config struct {
	Genesis   *core.Genesis `toml:",omitempty"`
	NetworkId uint64

	TxPool  core.TxPoolConfig
	Witness core.WitnessConfig

	// Adamnite DB options
	AdamniteDbCache   int
	AdamniteDbHandles int `toml:"-"`
}

var Defaults = Config{
	NetworkId: 888,
	TxPool:    core.DefaultTxPoolConfig,
	Witness:   core.DefaultWitnessConfig,

	AdamniteDbCache: 512,
}

func CreateConsensusEngine(node *node.Node, chainConfig *params.ChainConfig, db adamnitedb.Database) dpos.DPOS {
	engine := dpos.New(dpos.Config{})
	return engine
}
