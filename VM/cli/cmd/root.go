package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var root = &cobra.Command{
	Use: "adamnite",
	Short: "Adamnite is a CLI tool for compiling, executing and uploading smart contracts to the Adamnite Network",
}

func Execute() {
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	viper.SetDefault("author", "Adamnite Labs <dev@admanite.com>")
	viper.SetDefault("license", "apache")
}

func initConfig() {
}