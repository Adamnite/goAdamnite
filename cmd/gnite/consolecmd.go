package main

import "github.com/urfave/cli/v2"

var (
	consoleCommand = &cli.Command{
		Name:   "console",
		Usage:  "Start an interactive gnite console",
		Action: startConsole,
	}
)

func startConsole(ctx *cli.Context) error {
	return nil
}
