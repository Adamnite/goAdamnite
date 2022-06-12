package core

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
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
	WitnessList     []GenesisWitnessInfo `json:"witnessList"`
}

type GenesisWitnessInfo struct {
	address common.Address
	voters  []types.Voter
}

type GenesisAlloc map[common.Address]GenesisAccount

type GenesisAccount struct {
	PrivateKey []byte
	Balance    *big.Int
	Nonce      uint64
}

// Write writes the block and state of a genesis specification to the database.
func (g *Genesis) Write(db adamnitedb.Database) (*types.Block, error) {
	if db == nil {
		return nil, errors.New("db must be set")
	}

	statedb, _ := statedb.New(common.Hash{}, statedb.NewDatabase(db))
	for addr, account := range g.Alloc {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetNonce(addr, account.Nonce)
	}

	root := statedb.IntermediateRoot(false)

	head := &types.BlockHeader{
		ParentHash:      g.ParentHash,
		Time:            g.Time,
		Number:          new(big.Int).SetUint64(g.Number),
		Witness:         g.Witness,
		WitnessRoot:     common.HexToHash("0x8888888888888888888888888888888888"),
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

	rawdb.WriteEpochNumber(db, 0)
	rawdb.WriteBlock(db, block)

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
			common.HexToAddress("0x5d8124bb42734acb442b6992c73ecad2651612cd"): GenesisAccount{Balance: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(80000000000))},
			common.HexToAddress("0x5117dd7283175dfd686757784de62197bd2179a2"): GenesisAccount{Balance: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(80000000000))},
		},
		Witness: common.HexToAddress("0xada0000000000000000000000000000000000000"),
		WitnessList: []GenesisWitnessInfo{
			{
				address: common.HexToAddress("0x5d8124bb42734acb442b6992c73ecad2651612cd"),
				voters: []types.Voter{
					{
						Address:       common.HexToAddress("0x5117dd7283175dfd686757784de62197bd2179a2"),
						StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(100)),
					},
				},
			},
			{
				address: common.HexToAddress("0x5117dd7283175dfd686757784de62197bd2179a2"),
				voters: []types.Voter{
					{
						Address:       common.HexToAddress("0x5d8124bb42734acb442b6992c73ecad2651612cd"),
						StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(50)),
					},
				},
			},
		},
	}
}

func DefaultDemoGenesisBlock() *Genesis {
	return &Genesis{
		Config: params.DemoChainConfig,
		Alloc: GenesisAlloc{
			common.HexToAddress("0x5d8124bb42734acb442b6992c73ecad2651612cd"): GenesisAccount{Balance: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(80000000000))},
			common.HexToAddress("0x5117dd7283175dfd686757784de62197bd2179a2"): GenesisAccount{Balance: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(80000000000))},
		},
		Witness: common.HexToAddress("0xada0000000000000000000000000000000000000"),
		WitnessList: []GenesisWitnessInfo{
			{
				address: common.HexToAddress("0x5d8124bb42734acb442b6992c73ecad2651612cd"),
				voters: []types.Voter{
					{
						Address:       common.HexToAddress("0x5117dd7283175dfd686757784de62197bd2179a2"),
						StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(100)),
					},
				},
			},
			{
				address: common.HexToAddress("0x5117dd7283175dfd686757784de62197bd2179a2"),
				voters: []types.Voter{
					{
						Address:       common.HexToAddress("0x5d8124bb42734acb442b6992c73ecad2651612cd"),
						StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(50)),
					},
				},
			},
		},
	}
}

func WriteGenesisBlockWithOverride(db adamnitedb.Database, genesis *Genesis) (*params.ChainConfig, common.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return nil, common.Hash{}, errors.New("genesis has no chain configuration")
	}

	stored, err := rawdb.ReadHeaderHash(db, 0)
	if err != nil {
		return nil, common.Hash{}, errors.New("db access error")
	}

	if (stored == common.Hash{}) { // There is no genesis
		if genesis == nil {
			log15.Info("Writing mainnet genesis block")
			genesis = DefaultTestnetGenesisBlock()
		} else {
			log15.Info("Writing custom genesis block")
		}
		block, err := genesis.Write(db)
		if err != nil {
			return genesis.Config, common.Hash{}, err
		}
		return genesis.Config, block.Hash(), nil
	}

	return genesis.Config, stored, nil
}
