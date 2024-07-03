package main

import (
	"github.com/adamnite/go-adamnite/internal/flags"
	"github.com/urfave/cli/v2"
)

var (
	verbosityFlag = &cli.IntFlag{
		Name:     "verbosity",
		Usage:    "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=trace",
		Value:    3,
		Category: flags.LoggingCategory,
	}
)

var Flags = []cli.Flag{
	verbosityFlag,
}
