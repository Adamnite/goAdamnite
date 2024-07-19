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

	NetworkPort = &cli.IntFlag{
		Name:     "network.port",
		Usage:    "Adamnite p2p listening port",
		Value:    40908,
		Category: NetworkCateogry,
	}

	DataDir = &cli.StringFlag{
		Name:  "data-dir",
		Usage: "Adamnite data directory path",
	}
)

var (
	TestNetFlags = []cli.Flag{
		DeveloperFlag,
		TestnetFlag,
	}
	BlockNetFlags = append([]cli.Flag{MainnetFlag}, TestNetFlags...)

	NetworkFlags = []cli.Flag{
		NetworkPort,
	}

	BasicSettingsFlags = []cli.Flag{
		DataDir,
	}
)
