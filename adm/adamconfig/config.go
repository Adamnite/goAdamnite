package adamconfig

import "github.com/adamnite/go-adamnite/core"

type Config struct {
	Genesis   *core.Genesis `toml:",omitempty"`
	NetworkId uint64

	TxPool    core.TxPoolConfig
	Validator core.ValidatorConfig
}

var Defaults = Config{
	NetworkId: 888,
	TxPool:    core.DefaultTxPoolConfig,
}
