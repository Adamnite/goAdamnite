package flags

import (
	"os"
	"path/filepath"

	"github.com/adamnite/go-adamnite/params"
	"github.com/urfave/cli/v2"
)

// NewApp creates an app with sane defaults.
func NewApp(gitCommit, gitDate, usage string) *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Version = params.VersionWithCommit(gitCommit, gitDate)
	app.Usage = usage
	return app
}
