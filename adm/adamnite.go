package adm

import (
	"github.com/adamnite/go-adamnite/adm/adamconfig"
	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/node"
	"github.com/adamnite/go-adamnite/p2p"
)

// AdamniteImpl implements the Adamnite full node.
type AdamniteImpl struct {
	config *adamconfig.Config

	blockchain  *core.Blockchain
	txPool      *core.TxPool
	witnessPool *core.WitnessCandidatePool

	chainDB adamnitedb.Database

	p2pServer *p2p.Server

	dposEngine dpos.DPOS
}

func New(node *node.Node, config *adamconfig.Config) (*AdamniteImpl, error) {
	// ToDO: 1. Setup genesis block
	//       2. Create blockchain
	// 	     3. Create transaction pool
	//       4. Create witness pool

	// 1. Setup genesis block
	chainDB, err := node.OpenDatabase("adamnitedb", config.AdamniteDbCache, config.AdamniteDbHandles, false)
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, err := core.WriteGenesisBlockWithOverride(chainDB, config.Genesis)
	if err != nil {
		return nil, err
	}

	log15.Info("Initialised chain configuration", "config", chainConfig)
	log15.Info("Adamnite genesis hash", "hash", genesisHash)

	adamnite := &AdamniteImpl{
		config:     config,
		chainDB:    chainDB,
		dposEngine: adamconfig.CreateConsensusEngine(node, chainConfig, chainDB),
		p2pServer:  node.Server(),
	}

	adamnite.blockchain, err = core.NewBlockchain(chainDB, chainConfig, adamnite.dposEngine)
	if err != nil {
		return nil, err
	}

	adamnite.txPool = core.NewTxPool(config.TxPool, chainConfig, adamnite.blockchain)
	adamnite.witnessPool = core.NewWitnessPool(config.Witness, chainConfig, adamnite.blockchain)
	return adamnite, nil
}

func (adam *AdamniteImpl) DposEngine() dpos.DPOS { return adam.dposEngine }

func (adam *AdamniteImpl) StartConsensus() error {
	log15.Info("Consensus started")
	return nil
}
