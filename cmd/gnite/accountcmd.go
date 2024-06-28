package main

import "github.com/urfave/cli/v2"

var (
	accountCommand = &cli.Command{
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
	}
)
