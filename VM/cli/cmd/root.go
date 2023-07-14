package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Variables binded to Cobra commands
var (
	hexBytes     string
	filePath     string
	gas          uint64
	dbHost       string
	stateless    bool
	functionHash string
	functionArgs string
	testNet      bool
)

var rootCmd = &cobra.Command{
	Use:   "adamnite",
	Short: "Tool for compiling, executing and uploading smart contracts to the Adamnite blockchain",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	viper.SetDefault("author", "Adamnite Labs")
	viper.SetDefault("license", "MIT")
}
