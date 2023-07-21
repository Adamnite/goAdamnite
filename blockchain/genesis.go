package blockchain


// Probably move genesis block to somewhere else (like parameters). On that note we should probably remove blockchain in its entirety as well
import (
	"errors"
	"fmt"
	"math/big"
	//Replace with new DB/storage imports
)

type Genesis struct {
	Config          *params.ChainConfig  `json:"config"`
	Time            uint64               `json:"time"`
	Witness         utils.Address       `json:"witness"`
	Alloc           GenesisAlloc         `json:"alloc" gencodec:"required"`
	Number          uint64               `json:"number"`
	ParentHash      utils.Hash          `json:"parentHash"`
	Signature       utils.Hash          `json:"signature"`
	TransactionRoot utils.Hash          `json:"txroot"`
	WitnessList     []GenesisWitnessInfo `json:"witnessList"`
}


type Genesis struct {
	Config		*params.Config
	Time		uint64
	Witness		utils.Address //Set this to a locked addreess that cannot be accessed
	ParentHash				   
	Nonce
	Signature
	Allocation //Allocation Array, detailing addresses and balances for initial allocation
	GenesisChamber //Genesis Chamber A witnesses for the first round, round 0



}

type GenesisWitnessInfo struct {
	address utils.Address
	voters  []utils.Voter
}

type GenesisAlloc map[utils.Address]GenesisAccount

type GenesisAccount struct {
	PrivateKey []byte
	Balance    *big.Int
	Nonce      uint64
}

//Redo below code for new storage and originality



// Write writes the block and state of a genesis specification to the database.
func (g *Genesis) Write(db adamnitedb.Database) (*types.Block, error) {
	if db == nil {
		return nil, errors.New("db must be set")
	}

	statedb, _ := statedb.New(utils.Hash{}, statedb.NewDatabase(db))
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
		WitnessRoot:     utils.HexToHash("0x8888888888888888888888888888888888"),
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


//Make the allocations here
//Also have the predetermined witness list for the first round of consensus here
func DefaultTestnetGenesisBlock() *Genesis {
	return &Genesis{
		Config: params.TestnetChainConfig,
		Alloc: GenesisAlloc{
			utils.StringToAddress("3HCiFhyA1Kv3s25BeABHt7wW6N8y"): GenesisAccount{Balance: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(80000000000))},
			utils.StringToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"): GenesisAccount{Balance: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(80000000000))},
		},
		Witness: utils.StringToAddress("24oB2iyytkPa91zz6w8ywLfbSC2N"),
		WitnessList: []GenesisWitnessInfo{
			{
				address: utils.StringToAddress("3HCiFhyA1Kv3s25BeABHt7wW6N8y"),
				voters: []utils.Voter{
					{
						// From:          utils.StringToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"),
						StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(100)),
					},
				},
			},
			{
				address: utils.StringToAddress("0rbYLvW3xd9yEqpAhEBph4wPwFKo"),
				voters: []utils.Voter{
					{
						// From:          utils.StringToAddress("3HCiFhyA1Kv3s25BeABHt7wW6N8y"),
						StakingAmount: new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(50)),
					},
				},
			},
		},
	}
}

func WriteGenesisBlockWithOverride(db adamnitedb.Database, genesis *Genesis) (*params.ChainConfig, utils.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return nil, utils.Hash{}, errors.New("genesis has no chain configuration")
	}

	stored, err := rawdb.ReadHeaderHash(db, 0)
	if err != nil {
		return nil, utils.Hash{}, errors.New("db access error")
	}

	if (stored == utils.Hash{}) { // There is no genesis
		if genesis == nil {
			log15.Info("Writing testnet genesis block")
			genesis = DefaultTestnetGenesisBlock()
		} else {
			log15.Info("Writing custom genesis block")
		}
		block, err := genesis.Write(db)
		if err != nil {
			return genesis.Config, utils.Hash{}, err
		}
		return genesis.Config, block.Hash(), nil
	}

	return genesis.Config, stored, nil
}
