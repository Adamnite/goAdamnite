package launcher

import (
	"errors"
	"fmt"

	"github.com/adamnite/go-adamnite/adm"
	"github.com/adamnite/go-adamnite/internal/debug"
	"github.com/adamnite/go-adamnite/internal/flags"
	"github.com/adamnite/go-adamnite/internal/utils"
	"github.com/adamnite/go-adamnite/node"
	"github.com/adamnite/go-adamnite/params"
	"github.com/urfave/cli/v2"
)

var (
	gitCommit = ""
	gitDate   = ""

	app = flags.NewApp(gitCommit, gitDate, "the go-adamnite command line interface")

	networkFlags = []cli.Flag{
		&utils.NetworkIP,
		&utils.NATFlag,
	}
	witnessFlags = []cli.Flag{
		&utils.WitnessFalg,
		&utils.WitnessAddressFlag,
	}
)

func init() {
	app.Action = adamniteMain
	app.Version = params.VersionWithCommit(gitCommit, gitDate)
	app.HideVersion = true
	app.Commands = []*cli.Command{
		&accountCommand,
	}

	app.Flags = append(app.Flags, networkFlags...)
	app.Flags = append(app.Flags, debug.Flags...)
	app.Flags = append(app.Flags, witnessFlags...)

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

	node, adamnite := makeAdamniteNode(ctx)
	defer node.Close()

	startNode(ctx, node, adamnite)
	node.Wait()
	return nil
}

func startNode(ctx *cli.Context, node *node.Node, adamnite adm.AdamniteAPI) {
	utils.StartNode(ctx, node)

	if ctx.Bool(utils.WitnessFalg.Name) {
		if ctx.String(utils.WitnessAddressFlag.Name) == "" {
			utils.Fatalf("Witness address was not set")
		}

		adamniteImpl, ok := adamnite.(*adm.AdamniteImpl)

		if !ok {
			utils.Fatalf("Adamnite service not running: %v", errors.New("AdamniteImpl not created"))
		}

		if err := adamniteImpl.StartConsensus(); err != nil {
			utils.Fatalf("Failed to start consensus: %v", err)
		}
	}
}
