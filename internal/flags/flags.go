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

	// Validator
	ValidatorFlag = &cli.StringFlag{
		Name:     "validator.addr",
		Usage:    "gnite will be run as validator using this address",
		Category: DPOSCategory,
	}

	EpochDurationFlag = &cli.UintFlag{
		Name:     "dpos.epoch.duration",
		Usage:    "Adamnite DPOS epoch duration",
		Value:    360,
		Category: DPOSCategory,
	}

	MaxValidatorFlag = &cli.UintFlag{
		Name:     "dpos.validator.max",
		Usage:    "Adamnite DPOS epoch duration",
		Value:    10,
		Category: DPOSCategory,
	}

	BlockDurationFlag = &cli.UintFlag{
		Name:     "dpos.block.duration",
		Usage:    "Adamnite DPOS block duration",
		Value:    10,
		Category: DPOSCategory,
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

	ConsensusFlags = []cli.Flag{
		ValidatorFlag,
		EpochDurationFlag,
		MaxValidatorFlag,
		BlockDurationFlag,
	}
)
