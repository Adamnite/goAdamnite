package adamconfig

import (
	"time"
	"github.com/adamnite/go-adamnite/adm/validator"
	"github.com/adamnite/go-adamnite/dpos"

	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
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

	// Adamnite StateDB options
	AdamniteStateDBCash *trie.Config
}

var Defaults = Config{
	NetworkId:       888,
	TxPool:          core.DefaultTxPoolConfig,
	Witness:         dpos.DefaultWitnessConfig,

	Validator:       validator.Config{
		Recommit:  3000 * time.Millisecond,
	},
	AdamniteDbCache: 512,
	AdamniteStateDBCash: &trie.Config{
		Cache     : 256,     // Memory allowance (MB) to use for caching trie nodes in memory
		// Journal   : "", 	 // Journal of clean cache to survive node restarts
		// Preimages : false,   // Flag whether the preimage of trie key is recorded
	},
}

var DemoDefaults = Config{
	NetworkId: 890,
	TxPool:    core.DefaultTxPoolConfig,
	Witness:   dpos.DefaultDemoWitnessConfig,

	Validator: validator.Config{
		Recommit: 3000 * time.Millisecond,
	},

	AdamniteDbCache: 512,
	AdamniteStateDBCash: &trie.Config{
		Cache     : 256,     // Memory allowance (MB) to use for caching trie nodes in memory
		// Journal   : "", 	 // Journal of clean cache to survive node restarts
		// Preimages : false,   // Flag whether the preimage of trie key is recorded
	},
}
