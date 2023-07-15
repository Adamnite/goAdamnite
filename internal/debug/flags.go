package debug

import (
	"os"

	"github.com/urfave/cli/v2"
)

var (
	verbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail",
		Value: 3,
	}
)

var Flags = []cli.Flag{
	&verbosityFlag,
}

var glogger *log15.GlogHandler

func init() {
	glogger = log15.NewGlogHandler(log15.StreamHandler(os.Stderr, log15.TerminalFormat()))
	glogger.Verbosity(log15.LvlInfo)
	log15.Root().SetHandler(glogger)
}

func Setup(ctx *cli.Context) error {
	verbosity := ctx.Int(verbosityFlag.Name)
	glogger.Verbosity(log15.Lvl(verbosity))
	return nil
}
