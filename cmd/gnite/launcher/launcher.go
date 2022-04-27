package launcher

import (
	"fmt"

	"github.com/adamnite/go-adamnite/internal/debug"
	"github.com/adamnite/go-adamnite/internal/flags"
	"github.com/adamnite/go-adamnite/params"
	"github.com/urfave/cli/v2"
)

var (
	gitCommit = ""
	gitDate   = ""

	app = flags.NewApp(gitCommit, gitDate, "the go-adamnite command line interface")
)

func init() {
	app.Action = adamniteMain
	app.Version = params.VersionWithCommit(gitCommit, gitDate)
	app.HideVersion = true
	app.Commands = []*cli.Command{
		&accountCommand,
	}

	app.Before = func(ctx *cli.Context) error {
		if err := debug.Setup(ctx); err != nil {
			return err
		}

		return nil
	}

	app.After = func(ctx *cli.Context) error {
		return nil
	}
}

func Launch(args []string) error {
	return app.Run(args)
}

// main entry point into the Adamnite blockchain system.
func adamniteMain(ctx *cli.Context) error {
	if args := ctx.Args(); args.Len() > 0 {
		return fmt.Errorf("invalid command: %q", args.Get(0))
	}

	return nil
}
