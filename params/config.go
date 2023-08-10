package params

import "math/big"

var (
	TestnetChainConfig = &ChainConfig{
		ChainID: big.NewInt(889),
	}
)

type ChainConfig struct {
	ChainID *big.Int
}
