package blockchain

import (
	"math/big"
	"time"

	"github.com/adamnite/go-adamnite/internal/bargossip"
	"github.com/adamnite/go-adamnite/internal/blockchain/types"
	"github.com/adamnite/go-adamnite/internal/common"
	"github.com/adamnite/go-adamnite/log"
)

type CurrentState struct {
	validators []common.Address
	delegators []common.Address
}

type Config struct {
	MaxValidators uint
	EpochDuration uint
	BlockDuration uint
}

type Blockchain struct {
	Config

	genesisBlock  *types.Block
	db            []types.Block
	validatorAddr common.Address
	dataDir       string
	p2pServer     bargossip.LocalNode
	stateDB       CurrentState

	isSynced bool
}

func New(config Config, localNode bargossip.LocalNode, dataDir string, validatorAddr common.Address) Blockchain {
	bc := Blockchain{
		genesisBlock:  GetGenesisBlock(),
		db:            make([]types.Block, 0),
		validatorAddr: validatorAddr,
		dataDir:       dataDir,
		p2pServer:     localNode,
		isSynced:      false,
	}

	return bc
}

func (bc *Blockchain) Start() {
	go bc.StartSync()
	if bc.validatorAddr != common.HexToAddress("0x0000000000000000000000000000000000000000") {
		go bc.StartDPOS()
	}
}

func (bc *Blockchain) StartSync() {
	localBlockNum := bc.GetCurrentBlockNumber()
	remoteBlockNume := bc.p2pServer.GetPeersCurrentBlockNumber()

	bc.RefreshStateDB()

	if localBlockNum.Uint64() == remoteBlockNume.Uint64() {
		bc.isSynced = true
	} else if localBlockNum.Uint64() < remoteBlockNume.Uint64() {
		bc.isSynced = false
	} else {
		bc.isSynced = true
	}
}

func (bc *Blockchain) StartDPOS() {
	if bc.isSynced == true && bc.hasValidatorPermission() {
		log.Info("Validator starts", "Addr", bc.validatorAddr)

		for {
			if bc.getCurrentBlockValidator() == bc.validatorAddr {
				// TODO: Make Block
				time.Sleep(time.Duration(bc.BlockDuration * uint(time.Second)))
				header := types.Header{
					Number: bc.GetCurrentBlockNumber().Add(bc.GetCurrentBlockNumber(), big.NewInt(1)),
					Time:   uint64(time.Now().Unix()),
					Miner:  bc.validatorAddr,
				}

				txns := []types.Transaction{}

				block := types.NewBlock(header, txns)

				bc.p2pServer.BroadcastBlock(block)
			} else {
				// TODO: Validate Block
			}
		}
	}
}

func (bc *Blockchain) getCurrentBlockValidator() common.Address {
	nextBlockNum := bc.GetCurrentBlockNumber().Add(bc.GetCurrentBlockNumber(), big.NewInt(1))

	if nextBlockNum.Uint64()%360 == 0 {
		bc.RefreshStateDB()
		return bc.stateDB.validators[0]
	} else {
		index := nextBlockNum.Uint64() % uint64(len(bc.stateDB.validators))
		return bc.stateDB.validators[index]
	}
}

func (bc *Blockchain) hasValidatorPermission() bool {
	for _, val := range bc.stateDB.validators {
		if val == bc.validatorAddr {
			return true
		}
	}
	return false
}

func (bc *Blockchain) RefreshStateDB() {
	currentBlockNum := bc.GetCurrentBlockNumber()
	if currentBlockNum.Uint64() == 0 {
		for _, txn := range bc.genesisBlock.GetTransactions() {
			if txn.Type() == types.WitnessTxType {
				bc.stateDB.validators = append(bc.stateDB.validators, txn.GetValidator())
			}
		}
	}
}

func (bc *Blockchain) GetCurrentBlockNumber() *big.Int {
	lenDB := len(bc.db)

	if lenDB == 0 {
		return big.NewInt(0)
	} else {
		return bc.db[lenDB-1].GetBlockNumber()
	}
}
