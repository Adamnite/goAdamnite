package blockchain

import (
	"math/big"

	"github.com/adamnite/go-adamnite/internal/blockchain/types"
	"github.com/adamnite/go-adamnite/internal/common"
)

func GetGenesisBlock() *types.Block {
	header := types.Header{
		Number: big.NewInt(0),
		Time:   1722008731,
		Miner:  common.HexToAddress("0x0000000000000000000000000000000000000000"),
	}

	txns := []types.Transaction{
		*types.CreateNewWitnessTx(common.HexToAddress("0x0000000000000000000000000000000000000001")),
		*types.CreateNewWitnessTx(common.HexToAddress("0x0000000000000000000000000000000000000002")),
	}

	block := types.NewBlock(header, txns)
	return block
}
