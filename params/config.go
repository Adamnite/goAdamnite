package params

import "math/big"

var (
	MainnetChainConfig = &ChainConfig{
		ChainID: big.NewInt(888),
	}

	TestnetChainConfig = &ChainConfig{
		ChainID:     big.NewInt(889),
		NormalBlock: big.NewInt(12345),
	}

	DemoChainConfig = &ChainConfig{
		ChainID: big.NewInt(890),
	}
)

type ChainConfig struct {
	ChainID     *big.Int `json:"chainId"`
	NormalBlock *big.Int `json:"normalBlock,omitempty"`
}
