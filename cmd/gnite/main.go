package main

import (
	"fmt"
	"os"

	"github.com/adamnite/go-adamnite/internal/debug"
	"github.com/adamnite/go-adamnite/internal/utils"
	"github.com/adamnite/go-adamnite/log"
	"github.com/urfave/cli/v2"
)

func init() {

}

func main() {
	app := cli.NewApp()
	app.Name = "gnite"
	app.Usage = "CLI application for the go Adamnite node"
	app.Version = "1.0.0"

	// Define commands and flags
	app.Commands = []*cli.Command{
		accountCommand,
		consoleCommand,
	}

	app.Flags = []cli.Flag{}

	app.Flags = append(app.Flags, debug.Flags...)

	app.Before = func(ctx *cli.Context) error {
		utils.ShowBanner()

		if err := debug.Setup(ctx); err != nil {
			return err
		}

		return nil
	}

	app.Action = startGnite

	// Run the CLI application
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startGnite(ctx *cli.Context) error {
	log.Info("Adamnite gNite starting ...")
	return nil
}
