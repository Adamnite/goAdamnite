package launcher

import (
	"github.com/adamnite/go-adamnite/internal/flags"
	"github.com/urfave/cli/v2"
)

var (
	gitCommit = ""
	gitDate   = ""

	app = flags.NewApp(gitCommit, gitDate, "the go-adamnite command line interface")
)

func init() {
	app.Action = adamniteMain
}

func Launch(args []string) error {
	return app.Run(args)
}

// main entry point into the Adamnite blockchain system.
func adamniteMain(ctx *cli.Context) error {
	return nil
}
