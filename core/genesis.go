package core

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/params"
)

type Genesis struct {
	Config          *params.ChainConfig  `json:"config"`
	Time            uint64               `json:"time"`
	Witness         common.Address       `json:"witness"`
	Alloc           GenesisAlloc         `json:"alloc" gencodec:"required"`
	Number          uint64               `json:"number"`
	ParentHash      common.Hash          `json:"parentHash"`
	Signature       common.Hash          `json:"signature"`
	TransactionRoot common.Hash          `json:"txroot"`
	WitnessList     []types.Witness  `json:"witnessList"`
}

type GenesisAlloc map[common.Address]GenesisAccount

type GenesisAccount struct {
	PrivateKey []byte
	Balance    *big.Int
	Nonce      uint64
}

// Write writes the block and state of a genesis specification to the database.
func (g *Genesis) Write(db adamnitedb.Database, witnessConf dpos.WitnessConfig) (*types.Block, error) {
	if db == nil {
		return nil, errors.New("db must be set")
	}

	statedb, _ := statedb.New(common.Hash{}, statedb.NewDatabase(db))
	for addr, account := range g.Alloc {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetNonce(addr, account.Nonce)
	}

	root := statedb.IntermediateRoot(false)

	wp := dpos.NewWitnessPool(&witnessConf, g.Config, g.WitnessList)

	head := &types.BlockHeader{
		ParentHash:      g.ParentHash,
		Time:            g.Time,
		Number:          new(big.Int).SetUint64(g.Number),
		Witness:         g.Witness,
		WitnessRoot:     wp.RootHash(),
		Signature:       g.Signature,
		TransactionRoot: g.TransactionRoot,
		StateRoot:       root,
	}

	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, true, nil)

	block := types.NewBlock(head, nil, trie.NewStackTrie(nil))

	if block.Number().Sign() != 0 {
		return nil, fmt.Errorf("cannot write genesis block with blocknumber > 0")
	}

	config := g.Config
	if config == nil {
		return nil, fmt.Errorf("genesis config is not set")
	}

	rawdb.WriteHeaderHash(db, block.Header())
	rawdb.WriteEpochNumber(db, 0)
	rawdb.WriteBlock(db, block)
	rawdb.WriteCurrentBlockNumber(db, 0)
	
	wp.SaveWitnessPool(db)

	if wAddrs1, err := rawdb.ReadWitnessList(db, 0); err != nil {
		log15.Error("error to load witness list", "err", err, "wAddr", wAddrs1)
	}

	return block, nil
}

func DefaultGenesisBlock() *Genesis {
	return &Genesis{
		Config: params.MainnetChainConfig,
	}
}

func DefaultTestnetGenesisBlock() *Genesis {
	return &Genesis{
		Config: params.TestnetChainConfig,
		Alloc: GenesisAlloc{
			common.StringToAddress("3HCiFhyA1Kv3s25BeABHt7wW6N8y"): GenesisAccount{Balance: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(80000000000))},
			common.StringToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"): GenesisAccount{Balance: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(80000000000))},
		},
		Witness: common.StringToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"),
		WitnessList: []types.Witness{
			&types.WitnessImpl{
				Address: common.StringToAddress("347vh9kEJauu8LPBBKSiVES3GPag"),
				Voters: []types.Voter{
					{
						Address:       common.StringToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"),
						StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(100)),
					},
				},
			},
			&types.WitnessImpl{
				Address: common.StringToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"),
				Voters: []types.Voter{
					{
						Address:       common.StringToAddress("3HCiFhyA1Kv3s25BeABHt7wW6N8y"),
						StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(50)),
					},
				},
			},
		},
	}
}

// WriteGenesisBlockWithOverride writes the genesis block to local database.
func WriteGenesisBlockWithOverride(db adamnitedb.Database, genesis *Genesis, witnessConf dpos.WitnessConfig) (*params.ChainConfig, common.Hash, error) {
	if genesis == nil {
		return nil, common.Hash{}, errors.New("please set genesis information")
	} 

	if genesis != nil && genesis.Config == nil {
		return nil, common.Hash{}, errors.New("genesis has no chain configuration")
	}

	stored, err := rawdb.ReadHeaderHash(db, 0)
	if err != nil {
		if err == rawdb.ErrNoData {
			block, err := genesis.Write(db, witnessConf)
			if err != nil {
				log15.Error("Failed to write genesis block on db")
				return genesis.Config, common.Hash{}, err
			}
			log15.Info("Write genesis block on DB", "chain", genesis.Config)
			return genesis.Config, block.Hash(), nil
		} else {
			return nil, common.Hash{}, err
		}
	}

	return genesis.Config, stored, nil
}
