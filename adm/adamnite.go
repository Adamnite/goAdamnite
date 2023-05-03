package adm

import (
	"github.com/adamnite/go-adamnite/adm/adamconfig"
	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/protocols/adampro"
	"github.com/adamnite/go-adamnite/adm/validator"
	"github.com/adamnite/go-adamnite/bargossip"
	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/event"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/node"
)

// AdamniteImpl implements the Adamnite full node.
type AdamniteImpl struct {
	config *adamconfig.Config

	blockchain *core.Blockchain
	txPool     *core.TxPool

	witnessPool *dpos.WitnessPool

	handler *handler

	eventMux *event.TypeMux

	adamniteDialCandidates admnode.NodeIterator

	chainDB adamnitedb.Database

	p2pServer *bargossip.Server

	validator  *validator.Validator
	witness    common.Address

	// Adamnite consensus engine
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

	chainConfig, genesisHash, err := core.WriteGenesisBlockWithOverride(chainDB, config.Genesis, config.Witness)
	if err != nil {
		return nil, err
	}

	log15.Info("Initialised chain configuration", "config", chainConfig)
	log15.Info("Adamnite genesis hash", "hash", genesisHash.Hex())

	adamnite := &AdamniteImpl{
		config:     config,
		chainDB:    chainDB,
		dposEngine: dpos.New(chainConfig, chainDB),
		p2pServer:  node.Server(),
		eventMux:   node.EventMux(),
		witness:    config.Validator.WitnessAddress,
	}

	adamnite.blockchain, err = core.NewBlockchain(chainDB, chainConfig, adamnite.dposEngine, config.AdamniteStateDBCash)
	if err != nil {
		return nil, err
	}

	adamnite.txPool = core.NewTxPool(config.TxPool, chainConfig, adamnite.blockchain)
	adamnite.witnessPool, err = dpos.LoadWitnessPool(&config.Witness, chainConfig, chainDB)
	if err != nil {
		return nil, err
	}

	adamnite.handler, err = newHandler(&handlerParams{
		Database: chainDB,
		Chain:    adamnite.blockchain,
		ChainID:  adamnite.config.NetworkId,
		TxPool:   adamnite.txPool,
		EventMux: adamnite.eventMux,
	})
	if err != nil {
		return nil, err
	}

	adamnite.validator = validator.New(adamnite, &adamnite.config.Validator, chainConfig, adamnite.dposEngine, adamnite.eventMux)

	node.RegistProtocols(adamnite.Protocols())
	node.RegistServices(adamnite)

	return adamnite, nil
}

func (adam *AdamniteImpl) DposEngine() dpos.DPOS { return adam.dposEngine }

func (adam *AdamniteImpl) StartConsensus() error {
	go adam.validator.Start()
	return nil
}

func (adam *AdamniteImpl) Protocols() []bargossip.SubProtocol {
	return adampro.MakeProtocols(adam.handler, adam.config.NetworkId, adam.adamniteDialCandidates)
}

func (adam *AdamniteImpl) IsConnected() bool { return adam.p2pServer.IsConnected() }
func (adam *AdamniteImpl) Blockchain() *core.Blockchain   { return adam.blockchain }
func (adam *AdamniteImpl) TxPool() *core.TxPool           { return adam.txPool }
func (adam *AdamniteImpl) WitnessPool() *dpos.WitnessPool { return adam.witnessPool }
func (adam *AdamniteImpl) Witness() common.Address       { return adam.witness }

func (adam *AdamniteImpl) Start() error {
	adam.handler.Start(adam.p2pServer.MaxPendingConnections)
	return nil
}

func (adam *AdamniteImpl) Stop() error {
	// adam.adamniteDialCandidates.Close()
	adam.eventMux.Stop()
	return nil
}
