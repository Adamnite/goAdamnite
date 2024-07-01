package flags

import "github.com/urfave/cli/v2"

var (
	MainnetFlag = &cli.BoolFlag{
		Name:     "mainnet",
		Usage:    "Adamnite mainnet",
		Category: AdamniteCategory,
	}

	TestnetFlag = &cli.BoolFlag{
		Name:     "testnet",
		Usage:    "Adamnite testnet",
		Category: DevCategory,
	}

	// Dev mode
	DeveloperFlag = &cli.BoolFlag{
		Name:     "dev",
		Usage:    "Adamnite devnet for testing DPOS, bargossip",
		Category: DevCategory,
	}
)

var (
	TestNetFlags = []cli.Flag{
		DeveloperFlag,
		TestnetFlag,
	}
	NetworkFlags = append([]cli.Flag{MainnetFlag}, TestNetFlags...)
)
