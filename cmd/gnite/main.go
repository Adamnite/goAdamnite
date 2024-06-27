package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func init() {

}

func main() {
	// Initialize logrus logger
	logger := log.New()

	logger.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.InfoLevel)

	app := cli.NewApp()
	app.Name = "gnite"
	app.Usage = "CLI application for the go Adamnite node"
	app.Version = "1.0.0"

	// Define commands and flags
	app.Commands = []*cli.Command{
		{
			Name:  "account",
			Usage: "Greet someone",
			Subcommands: []*cli.Command{
				{
					Name:  "new",
					Usage: "Create a new Adamnite account",

					Action: func(c *cli.Context) error {
						return nil
					},
				},
				{
					Name:  "list",
					Usage: "Print summary of existing accounts",

					Action: func(c *cli.Context) error {
						return nil
					},
				},
				{
					Name:  "import",
					Usage: "Import a private key into a new account",

					Action: func(c *cli.Context) error {
						return nil
					},
				},
			},
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	}

	app.Flags = []cli.Flag{}

	app.Flags = append(app.Flags, []cli.Flag{
		&cli.IntFlag{
			Name:  "verbosity",
			Usage: "",
			Action: func(ctx *cli.Context, i int) error {
				switch i {
				case 0:
					logger.SetLevel(log.PanicLevel)
				case 1:
					logger.SetLevel(log.FatalLevel)
				case 2:
					logger.SetLevel(log.ErrorLevel)
				case 3:
					logger.SetLevel(log.WarnLevel)
				case 4:
					logger.SetLevel(log.InfoLevel)
				case 5:
					logger.SetLevel(log.DebugLevel)
				case 6:
					logger.SetLevel(log.TraceLevel)
				default:
					logger.SetLevel(log.InfoLevel)
				}

				return nil
			},
		},
	}...)

	app.Action = func(ctx *cli.Context) error {
		logger.WithFields(log.Fields{"AAA": "2324234234"}).Info("Starting the gnite Node!!!")
		logger.Info("Starting the gnite Node!!!")
		return nil
	}

	// Run the CLI application
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
