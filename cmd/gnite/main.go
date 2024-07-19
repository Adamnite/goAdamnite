package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/adamnite/go-adamnite/internal/debug"
	"github.com/adamnite/go-adamnite/internal/flags"
	"github.com/adamnite/go-adamnite/internal/utils"
	"github.com/adamnite/go-adamnite/log"
	"github.com/urfave/cli/v2"
)

var (
	nodeFlags = []cli.Flag{}
)

func init() {
	nodeFlags = append(nodeFlags, flags.NetworkFlags...)
	nodeFlags = append(nodeFlags, flags.BlockNetFlags...)
	nodeFlags = append(nodeFlags, flags.BasicSettingsFlags...)
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

	app.Flags = flags.Merge(app.Flags, debug.Flags, nodeFlags)

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

	if args := ctx.Args().Slice(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}

	prepare(ctx)

	node := makeGniteNode(ctx)
	defer node.Close()

	node.Start()

	go func() {
		// Create a channel to receive OS signals
		sigCh := make(chan os.Signal, 1)

		// Notify the signal channel on SIGINT (Ctrl+C) and SIGTERM (termination)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// Block until a signal is received
		sig := <-sigCh

		log.Info("Signal received", "SIG", sig)

		node.Close()

		// Exit the program
		// os.Exit(0)
	}()

	node.Wait()
	return nil
}

func prepare(ctx *cli.Context) {
	switch {
	case ctx.IsSet(flags.DeveloperFlag.Name):
		log.Info("Starting gNite on Adamnite devnet ...")
		log.Debug("All blockchain datas will be stored on the memory db.")
	case ctx.IsSet(flags.MainnetFlag.Name):
		log.Info("Starting gNite on Adamnite mainnet ...")
	case ctx.IsSet(flags.TestnetFlag.Name):
		log.Info("Starting gNite on Adamnite testnet ...")
	default:
		log.Info("Starting gNite on Adamnite mainnet ...")
	}
}
