package adamconfig

import (
	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/validator"
	"github.com/adamnite/go-adamnite/dpos"

	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/node"
	"github.com/adamnite/go-adamnite/params"
)

type Config struct {
	Genesis   *core.Genesis `toml:",omitempty"`
	NetworkId uint64

	TxPool  core.TxPoolConfig
	Witness dpos.WitnessConfig

	Validator validator.Config

	// Adamnite DB options
	AdamniteDbCache   int
	AdamniteDbHandles int `toml:"-"`
}

var Defaults = Config{
	NetworkId:       888,
	TxPool:          core.DefaultTxPoolConfig,
	Witness:         dpos.DefaultWitnessConfig,
	Validator:       validator.DefaultConfig,
	AdamniteDbCache: 512,
}

var DemoDefaults = Config{
	NetworkId: 890,
	TxPool:    core.DefaultTxPoolConfig,
	Witness:   dpos.DefaultDemoWitnessConfig,
	Validator: validator.DefaultDemoConfig,

	AdamniteDbCache: 512,
}

func CreateConsensusEngine(node *node.Node, chainConfig *params.ChainConfig, db adamnitedb.Database) dpos.Engine {
	engine := dpos.New(dpos.Config{}, db)

	return engine
}
